package synchandler

import (
	"context"
	"fmt"

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
	KeyperSetManagerContractABI, err = bindings.KeyperSetManagerMetaData.GetAbi()
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
		ethClient:               client,
		keyperSetManager:        ksm,
	})
}

type KeyperSetAdded struct {
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
	query syncer.ChainUpdateContext,
	events []bindings.KeyperSetManagerKeyperSetAdded,
) error {
	// TODO: we don't handle reorgs here.
	// This is because we don't have a good way to deal with
	// them:
	// When we originally insert the event, we would have to save
	// the insert block-hash in the db and upon a reorg delete the keypersets
	// by insert block-hash. We can't do this in production, since we don't
	// have a good database migration strategy and framework yet.

	for _, ev := range events {
		ks, err := QueryFullKeyperSetFromKeyperSetAddedEvent(ctx, handler.ethClient, ev, handler.keyperSetManager)
		if err != nil {
			// TODO: logging
			// handler.log.Error("KeyperSetAdded event, error querying keyperset-data", "error", err)
		}
		err = handler.processNewKeyperSet(ctx, ks)
		if err != nil {
			// TODO: logging
			// handler.log.Error("KeyperSetAdded event, error writing to database", "error", err)
		}
	}
	return nil
}

func (handler *KeyperSetAdded) processNewKeyperSet(ctx context.Context, ev *KeyperSet) error {
	isMember := false
	for _, m := range ev.Members {
		if m.Cmp(handler.ethereumAddress) == 0 {
			isMember = true
			break
		}
	}
	_ = isMember
	// FIXME: use the old zerologger again
	// handler.log.Info(
	// 	"new keyper set added",
	// 	"activation-block", ev.ActivationBlock,
	// 	"eon", ev.Eon,
	// 	"num-members", len(ev.Members),
	// 	"threshold", ev.Threshold,
	// 	"is-member", isMember,
	// )

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

type KeyperSet struct {
	ActivationBlock uint64
	Members         []common.Address
	Threshold       uint64
	Eon             uint64

	AtBlockNumber *number.BlockNumber
}

// QueryFullKeyperSetFromKeyperSetAddedEvent polls some additional
// data from the contracts in order to construct the full set of
// information for a keyper-set.
// This has to be done because not all information relevant to
// the keyperset is included in the KeyperSetAdded event.
func QueryFullKeyperSetFromKeyperSetAddedEvent(
	ctx context.Context,
	ethClient client.Client,
	event bindings.KeyperSetManagerKeyperSetAdded,
	keyperSetManager *bindings.KeyperSetManager,
) (*KeyperSet, error) {
	keyperSet, err := bindings.NewKeyperSet(event.KeyperSetContract, ethClient)
	if err != nil {
		return nil, fmt.Errorf("can't bind KeyperSet contract", err)
	}
	opts := &bind.CallOpts{
		BlockHash: event.Raw.BlockHash,
		Context:   ctx,
	}
	// the manager only accepts final keyper sets,
	// so we expect this to be final now.
	final, err := keyperSet.IsFinalized(opts)
	if err != nil {
		return nil, makeCallError("IsFinalized", err)
	}
	if !final {
		return nil, errors.New("contract did accept unfinalized keyper-sets")
	}
	members, err := keyperSet.GetMembers(opts)
	if err != nil {
		return nil, makeCallError("Members", err)
	}
	threshold, err := keyperSet.GetThreshold(opts)
	if err != nil {
		return nil, makeCallError("Threshold", err)
	}
	eon, err := keyperSetManager.GetKeyperSetIndexByBlock(opts, event.ActivationBlock)
	if err != nil {
		return nil, makeCallError("KeyperSetIndexByBlock", err)
	}
	return &KeyperSet{
		ActivationBlock: event.ActivationBlock,
		Members:         members,
		Threshold:       threshold,
		Eon:             eon,
		AtBlockNumber:   number.BigToBlockNumber(opts.BlockNumber),
	}, nil
}
