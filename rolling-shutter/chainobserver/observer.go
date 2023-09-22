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

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb"
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
	contracts     *deployment.Contracts
	dbpool        *pgxpool.Pool
	eventHandlers map[reflect.Type]EventHandlerFunc
}

func MakeHandler[T any](handler EventHandlerFuncGeneric[T]) EventHandlerFunc {
	anyHandler := func(ctx context.Context, tx pgx.Tx, anyEvent any) error {
		event, ok := anyEvent.(T)
		if !ok {
			// TODO better error message
			return errors.New("type mismatch")
		}
		return handler(ctx, tx, event)
	}
	return anyHandler
}

func New(contracts *deployment.Contracts, dbpool *pgxpool.Pool) *ChainObserver {
	return &ChainObserver{contracts: contracts, dbpool: dbpool, eventHandlers: make(map[reflect.Type]EventHandlerFunc)}
}

func (chainobs *ChainObserver) Observe(ctx context.Context, events map[*eventsyncer.EventType]EventHandlerFunc) error {
	db := chainobsdb.New(chainobs.dbpool)
	eventSyncProgress, err := db.GetEventSyncProgress(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get last synced event from db")
	}

	eventTypes := []*eventsyncer.EventType{}
	for eventType, handler := range events {
		eventTypes = append(eventTypes, eventType)
		chainobs.eventHandlers[eventType.Type] = handler
	}

	var fromBlock, fromLogIndex uint64
	if len(events) == 0 {
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

type (
	EventHandlerFunc               func(context.Context, pgx.Tx, any) error
	EventHandlerFuncGeneric[T any] func(context.Context, pgx.Tx, T) error
)

func (chainobs *ChainObserver) handleEvent(
	ctx context.Context, tx pgx.Tx, event interface{},
) error {
	// FIXME indirect etc?
	eventType := reflect.TypeOf(event)
	handler, ok := chainobs.eventHandlers[eventType]
	if !ok {
		log.Info().Str("event-type", reflect.TypeOf(event).String()).Interface("event", event).
			Msg("ignoring unknown event")
		return nil
	}
	err := handler(ctx, tx, event)
	if err != nil {
		log.Error().Err(err).Str("event", eventType.Name()).Msg("error during handler invocation")
	}
	return nil
}

// handleEventSyncUpdate handles events and advances the sync state, but rolls back any db updates
// on failure.
func (chainobs *ChainObserver) handleEventSyncUpdate(
	ctx context.Context, eventSyncUpdate eventsyncer.EventSyncUpdate,
) error {
	return chainobs.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		if eventSyncUpdate.Event != nil {
			if err := chainobs.handleEvent(ctx, tx, eventSyncUpdate.Event); err != nil {
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
		db := chainobsdb.New(tx)
		if err := db.UpdateEventSyncProgress(ctx, chainobsdb.UpdateEventSyncProgressParams{
			NextBlockNumber: int32(nextBlockNumber),
			NextLogIndex:    int32(nextLogIndex),
		}); err != nil {
			return errors.Wrap(err, "failed to update last synced event")
		}
		return nil
	})
}

func (chainobs *ChainObserver) handleKeypersConfigsListNewConfigEvent(
	ctx context.Context, tx pgx.Tx, event contract.KeypersConfigsListNewConfig,
) error {
	addrs, err := retryGetAddrs(ctx, chainobs.contracts.Keypers, event.KeyperSetIndex)
	if err != nil {
		return err
	}
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
	db := chainobsdb.New(tx)
	err = db.InsertKeyperSet(ctx, chainobsdb.InsertKeyperSetParams{
		KeyperConfigIndex:     int64(event.KeyperConfigIndex),
		ActivationBlockNumber: int64(event.ActivationBlockNumber),
		Keypers:               shdb.EncodeAddresses(addrs),
		Threshold:             int32(event.Threshold),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to insert keyper set into db")
	}
	return nil
}

// TODO move in separate handler
func (chainobs *ChainObserver) handleCollatorConfigsListNewConfigEvent(
	ctx context.Context, db *chainobsdb.Queries, event contract.CollatorConfigsListNewConfig,
) error {
	addrs, err := retryGetAddrs(ctx, chainobs.contracts.Collators, event.CollatorSetIndex)
	if err != nil {
		return err
	}
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
	if len(addrs) > 1 {
		return errors.Errorf("got multiple collators from collator addrs set contract: %s", addrs)
	} else if len(addrs) == 1 {
		err := db.InsertChainCollator(ctx, chainobsdb.InsertChainCollatorParams{
			ActivationBlockNumber: int64(event.ActivationBlockNumber),
			Collator:              shdb.EncodeAddress(addrs[0]),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to insert collator into db")
		}
	}
	return nil
}
