package chainobserver

import (
	"context"
	"math"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	cltrdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/collator"
	kprdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	syncdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/sync"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

const finalityOffset = 3

func retryGetAddrs(ctx context.Context, addrsSeq *contract.AddrsSeq, n uint64) ([]common.Address, error) {
	callOpts := &bind.CallOpts{
		Pending: false,
		// We call for the current height instead of the height at which the event was emitted,
		// because the sets cannot change retroactively and we won't need an archive node.
		BlockNumber: nil,
		Context:     ctx,
	}
	addrs, err := retry.FunctionCall(ctx, func(_ context.Context) ([]common.Address, error) {
		return addrsSeq.GetAddrs(callOpts, n)
	})
	if err != nil {
		return []common.Address{}, errors.Wrapf(err, "failed to query address set from contract")
	}
	return addrs, nil
}

type ChainObserver struct {
	contracts *deployment.Contracts
	dbpool    *pgxpool.Pool
}

func New(contracts *deployment.Contracts, dbpool *pgxpool.Pool) *ChainObserver {
	return &ChainObserver{contracts: contracts, dbpool: dbpool}
}

func (chainobs *ChainObserver) Observe(ctx context.Context, eventTypes []*eventsyncer.EventType) error {
	db := syncdb.New(chainobs.dbpool)
	eventSyncProgress, err := db.GetEventSyncProgress(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get last synced event from db")
	}

	var fromBlock, fromLogIndex uint64
	if len(eventTypes) == 0 {
		return errors.New("no events to observe")
	}

	// first find the min of all event's from-blocks
	fromBlock = eventTypes[0].FromBlockNumber
	for _, event := range eventTypes {
		if event.FromBlockNumber < fromBlock {
			fromBlock = event.FromBlockNumber
		}
	}

	// then check if our saved progress is already later
	progressBlock := uint64(eventSyncProgress.NextBlockNumber)
	if progressBlock > fromBlock {
		fromBlock = progressBlock
		// only use the saved log index when we're using the
		// saved block-number
		fromLogIndex = uint64(eventSyncProgress.NextLogIndex)
	}

	log.Info().Uint64("from-block", fromBlock).Uint64("from-log-index", fromLogIndex).
		Msg("starting event syncing")
	syncer := eventsyncer.New(chainobs.contracts.Client, finalityOffset, eventTypes, fromBlock, fromLogIndex)

	errorgroup, errorctx := errgroup.WithContext(ctx)
	errorgroup.Go(func() error {
		return syncer.Run(errorctx)
	})
	errorgroup.Go(func() error {
		for {
			select {
			case <-errorctx.Done():
				return errorctx.Err()
			default:
				eventSyncUpdate, err := syncer.Next(errorctx)
				if err != nil {
					return err
				}
				if err := chainobs.handleEventSyncUpdate(errorctx, eventSyncUpdate); err != nil {
					return err
				}
			}
		}
	})
	return errorgroup.Wait()
}

type newKeyperConfig struct {
	contract.KeypersConfigsListNewConfig
	addrs []common.Address
}

type newCollatorConfig struct {
	contract.CollatorConfigsListNewConfig
	addrs []common.Address
}

func (chainobs *ChainObserver) amendEvent(
	ctx context.Context, event interface{},
) (interface{}, error) {
	switch event := event.(type) {
	case contract.KeypersConfigsListNewConfig:
		addrs, err := retryGetAddrs(ctx, chainobs.contracts.Keypers, event.KeyperSetIndex)
		if err != nil {
			return nil, err
		}
		return newKeyperConfig{KeypersConfigsListNewConfig: event, addrs: addrs}, nil
	case contract.CollatorConfigsListNewConfig:
		addrs, err := retryGetAddrs(ctx, chainobs.contracts.Collators, event.CollatorSetIndex)
		if err != nil {
			return nil, err
		}
		return newCollatorConfig{CollatorConfigsListNewConfig: event, addrs: addrs}, nil
	}
	return event, nil
}

// handleEventSyncUpdate handles events and advances the sync state, but rolls back any db updates
// on failure.
func (chainobs *ChainObserver) handleEventSyncUpdate(
	ctx context.Context, eventSyncUpdate eventsyncer.EventSyncUpdate,
) error {
	var err error
	eventSyncUpdate.Event, err = chainobs.amendEvent(ctx, eventSyncUpdate.Event)
	if err != nil {
		return err
	}
	return chainobs.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		kdb := kprdb.New(tx)
		cltdb := cltrdb.New(tx)
		sncdb := syncdb.New(tx)

		if eventSyncUpdate.Event != nil {
			if err := chainobs.handleEvent(ctx, kdb, cltdb, eventSyncUpdate.Event); err != nil {
				return err
			}
		}

		var nextBlockNumber uint64
		var nextLogIndex uint64
		if eventSyncUpdate.Event == nil {
			nextBlockNumber = eventSyncUpdate.BlockNumber + 1
			nextLogIndex = 0
		} else {
			nextBlockNumber = eventSyncUpdate.BlockNumber
			nextLogIndex = eventSyncUpdate.LogIndex + 1
		}
		if err := sncdb.UpdateEventSyncProgress(ctx, syncdb.UpdateEventSyncProgressParams{
			NextBlockNumber: int32(nextBlockNumber),
			NextLogIndex:    int32(nextLogIndex),
		}); err != nil {
			return errors.Wrap(err, "failed to update last synced event")
		}
		return nil
	})
}

func (chainobs *ChainObserver) handleEvent(
	ctx context.Context, kdb *kprdb.Queries, cltdb *cltrdb.Queries, event interface{},
) error {
	var err error
	switch event := event.(type) {
	case newKeyperConfig:
		err = chainobs.handleKeypersConfigsListNewConfigEvent(ctx, kdb, event)
	case newCollatorConfig:
		err = chainobs.handleCollatorConfigsListNewConfigEvent(ctx, cltdb, event)
	default:
		log.Info().Str("event-type", reflect.TypeOf(event).String()).Interface("event", event).
			Msg("ignoring unknown event")
	}
	return err
}

func (chainobs *ChainObserver) handleKeypersConfigsListNewConfigEvent(
	ctx context.Context, db *kprdb.Queries, event newKeyperConfig,
) error {
	log.Info().
		Uint64("block-number", event.Raw.BlockNumber).
		Uint64("keyper-config-index", event.KeyperConfigIndex).
		Uint64("activation-block-number", event.ActivationBlockNumber).
		Msg("handling NewConfig event from keypers config contract")

	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf(
			"activation block number %d from config contract would overflow int64",
			event.ActivationBlockNumber)
	}
	err := db.InsertKeyperSet(ctx, kprdb.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(event.KeyperConfigIndex),
		ActivationBlockNumber: int64(event.ActivationBlockNumber),
		Keypers:               shdb.EncodeAddresses(event.addrs),
		Threshold:             int32(event.Threshold),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to insert keyper set into db")
	}
	return nil
}

func (chainobs *ChainObserver) handleCollatorConfigsListNewConfigEvent(
	ctx context.Context, db *cltrdb.Queries, event newCollatorConfig,
) error {
	log.Info().
		Uint64("block-number", event.Raw.BlockNumber).
		Uint64("collator-config-index", event.CollatorConfigIndex).
		Uint64("activation-block-number", event.ActivationBlockNumber).
		Msg("handling NewConfig event from collator config contract")
	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf(
			"activation block number %d from config contract would overflow int64",
			event.ActivationBlockNumber,
		)
	}
	if len(event.addrs) > 1 {
		return errors.Errorf("got multiple collators from collator addrs set contract: %s", event.addrs)
	} else if len(event.addrs) == 1 {
		err := db.InsertChainCollator(ctx, cltrdb.InsertChainCollatorParams{
			ActivationBlockNumber: int64(event.ActivationBlockNumber),
			Collator:              shdb.EncodeAddress(event.addrs[0]),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to insert collator into db")
		}
	}
	return nil
}
