package synchandler

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

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

var ErrParseKeyperSet = errors.New("can't parse KeyperSet")

func makeCallError(attrName string, err error) error {
	return fmt.Errorf("could not retrieve `%s` from contract: %w", attrName, err)
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
	ethClient client.Client,
	contractAddress,
	ethereumAddress common.Address,
) (syncer.ContractEventHandler, error) {
	ksm, err := bindings.NewKeyperSetManager(contractAddress, ethClient)
	if err != nil {
		return nil, err
	}
	return syncer.WrapHandler(&KeyperSetAdded{
		evABI:                   KeyperSetManagerContractABI,
		keyperSetManagerAddress: contractAddress,
		ethereumAddress:         ethereumAddress,
		dbpool:                  db,
		ethClient:               ethClient,
		keyperSetManager:        ksm,
	})
}

type KeyperSetAdded struct {
	evABI                   *abi.ABI
	keyperSetManagerAddress common.Address
	// our address to check wether we are part of the keyper-set
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

func (*KeyperSetAdded) Event() string {
	return "KeyperSetAdded"
}

func (handler *KeyperSetAdded) ABI() abi.ABI {
	return *handler.evABI
}

func (handler *KeyperSetAdded) Accept(
	_ context.Context,
	_ types.Header,
	_ bindings.KeyperSetManagerKeyperSetAdded,
) (bool, error) {
	return true, nil
}

func (handler *KeyperSetAdded) Handle(
	ctx context.Context,
	update syncer.ChainUpdateContext,
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
			log.Error().
				Err(err).
				Msg("KeyperSetAdded event, error querying keyperset-data")
		}
		err = handler.processNewKeyperSet(ctx, ks)
		if err != nil {
			log.Error().Err(err).Msg("KeyperSetAdded event, error writing to database")
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
	log.Info().
		Uint64("activation-block", ev.ActivationBlock).
		Uint64("eon", ev.Eon).
		Int("num-members", len(ev.Members)).
		Uint64("threshold", ev.Threshold).
		Bool("is-member", isMember).
		Msg("new keyper set added")

	// TODO: before, we were notifying the SequencerTransactionSubmitted
	// handler when we were part of the newly inserted KeyperSet.
	// This was an optimisation measure to not let the node
	// insert SequencerTransactionSubmitted events into it's DB for
	// Eons that are before the point where it is part of the KeyperSet.
	// This optimisation is for now omitted, since it unnecessarily coupled
	// both handler. It was mainly done like this to avoid adding an
	// additional field 'weAreMember' to the KeyperSet in the db.
	// XXX: do we have to insert keypersets we are not part of?
	// Maybe it would be easiest to just ignore those keypersets?

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

		// we insert the keyperset into the db, even though we are not member of it.
		// Since there is no field to mark our keypersets, we would always
		// have to iterate over all keypersets...
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
		return nil, fmt.Errorf("can't bind KeyperSet contract: %w", err)
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
