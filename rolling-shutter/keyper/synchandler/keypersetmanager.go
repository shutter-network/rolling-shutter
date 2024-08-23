package synchandler

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/shop-contracts/bindings"
)

var (
	ErrParseKeyperSet = errors.New("can't parse KeyperSet")

	errNilCallOpts        = errors.New("nil call-opts")
	errNilOptsBlockNumber = errors.New("opts block-number is nil, but 'latest' not allowed")
	errLatestBlock        = errors.New("'nil' latest block")
)

func makeCallError(attrName string, err error) error {
	return errors.Wrapf(err, "could not retrieve `%s` from contract", attrName)
}

func init() {
	var err error
	KeyperSetManagerContractABI, err = bindings.KeyBroadcastContractMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

var KeyperSetManagerContractABI *abi.ABI

func NewKeyperSetAdded(
	db *pgxpool.Pool,
	client client.Client,
	contractAddress,
	ethereumAddress common.Address,
) (syncer.ContractEventHandler, error) {
	ksm, err := bindings.NewKeyperSetManager(contractAddress, client)
	if err != nil {
		return nil, err
	}
	return syncer.WrapHandler(&KeyperSetAdded{
		// TODO: log
		// log:                     log,
		evABI:                   KeyperSetManagerContractABI,
		keyperSetManagerAddress: contractAddress,
		ethereumAddress:         ethereumAddress,
		dbpool:                  db,
		keyperSetManager:        ksm,
	})
}

type KeyperSetAdded struct {
	log                     log.Logger
	evABI                   *abi.ABI
	keyperSetManagerAddress common.Address
	// own address
	ethereumAddress common.Address
	dbpool          *pgxpool.Pool

	// we only need this because we have to poll
	// additional data from the contract
	ethClient        client.Client
	keyperSetManager *bindings.KeyperSetManager
}

func (handler *KeyperSetAdded) Address() common.Address {
	return handler.keyperSetManagerAddress
}

func (handler *KeyperSetAdded) Log(msg string, ctx ...any) {
	handler.log.Info(msg, ctx)
}

func (_ *KeyperSetAdded) Event() string {
	return "KeyperSetAdded"
}

func (handler *KeyperSetAdded) ABI() abi.ABI {
	return *handler.evABI
}
func (handler *KeyperSetAdded) Accept(
	ctx context.Context,
	header types.Header,
	ev bindings.KeyperSetManagerKeyperSetAdded,
) (bool, error) {
	return true, nil
}
func (handler *KeyperSetAdded) Handle(
	ctx context.Context,
	query syncer.QueryContext,
	events []bindings.KeyperSetManagerKeyperSetAdded,
) error {
	// TODO: handle qCtx.Remove, if possible?
	for _, ev := range events {
		opts := &bind.CallOpts{
			BlockHash: ev.Raw.BlockHash,
			Context:   ctx,
		}
		ks, err := handler.queryKeyperSetData(opts, ev.KeyperSetContract, ev.ActivationBlock)
		if err != nil {
			handler.log.Error("KeyperSetAdded event, error querying keyperset-data", "error", err)
		}
		err = handler.processNewKeyperSet(ctx, ks)
		if err != nil {
			handler.log.Error("KeyperSetAdded event, error writing to database", "error", err)
		}
	}
	return nil
}

type keyperSet struct {
	ActivationBlock uint64
	Members         []common.Address
	Threshold       uint64
	Eon             uint64

	AtBlockNumber *number.BlockNumber
}

// NOTE: unfortunately we have to poll
// some data from the blockchain here, because they are not all
// available from the event direcly.
func (handler *KeyperSetAdded) queryKeyperSetData(
	opts *bind.CallOpts,
	keyperSetContract common.Address,
	activationBlock uint64,
) (*keyperSet, error) {
	ks, err := bindings.NewKeyperSet(keyperSetContract, handler.ethClient)
	if err != nil {
		return nil, errors.Wrap(err, "could not bind to KeyperSet contract")
	}
	// the manager only accepts final keyper sets,
	// so we expect this to be final now.
	final, err := ks.IsFinalized(opts)
	if err != nil {
		return nil, makeCallError("IsFinalized", err)
	}
	if !final {
		return nil, errors.New("contract did accept unfinalized keyper-sets")
	}
	members, err := ks.GetMembers(opts)
	if err != nil {
		return nil, makeCallError("Members", err)
	}
	threshold, err := ks.GetThreshold(opts)
	if err != nil {
		return nil, makeCallError("Threshold", err)
	}
	eon, err := handler.keyperSetManager.GetKeyperSetIndexByBlock(opts, activationBlock)
	if err != nil {
		return nil, makeCallError("KeyperSetIndexByBlock", err)
	}
	return &keyperSet{
		ActivationBlock: activationBlock,
		Members:         members,
		Threshold:       threshold,
		Eon:             eon,
		AtBlockNumber:   number.BigToBlockNumber(opts.BlockNumber),
	}, nil
}

func (handler *KeyperSetAdded) processNewKeyperSet(ctx context.Context, ev *keyperSet) error {
	isMember := false
	for _, m := range ev.Members {
		if m.Cmp(handler.ethereumAddress) == 0 {
			isMember = true
			break
		}
	}
	handler.log.Info(
		"new keyper set added",
		"activation-block", ev.ActivationBlock,
		"eon", ev.Eon,
		"num-members", len(ev.Members),
		"threshold", ev.Threshold,
		"is-member", isMember,
	)

	return handler.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		obskeyperdb := obskeyper.New(tx)

		keyperConfigIndex, err := medley.Uint64ToInt64Safe(ev.Eon)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		activationBlockNumber, err := medley.Uint64ToInt64Safe(ev.ActivationBlock)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		threshold, err := medley.Uint64ToInt64Safe(ev.Threshold)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}

		return obskeyperdb.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
			KeyperConfigIndex:     keyperConfigIndex,
			ActivationBlockNumber: activationBlockNumber,
			Keypers:               shdb.EncodeAddresses(ev.Members),
			Threshold:             int32(threshold),
		})
	})
}
