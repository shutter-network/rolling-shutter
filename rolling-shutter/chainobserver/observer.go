package chainobserver

import (
	"context"
	"log"
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shuttermint/commondb"
	"github.com/shutter-network/shutter/shuttermint/contract"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/medley/eventsyncer"
	"github.com/shutter-network/shutter/shuttermint/shdb"
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
	addrs, err := medley.Retry(ctx, func() ([]common.Address, error) {
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
	db := commondb.New(chainobs.dbpool)
	eventSyncProgress, err := db.GetEventSyncProgress(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get last synced event from db")
	}
	fromBlock := uint64(eventSyncProgress.NextBlockNumber)
	fromLogIndex := uint64(eventSyncProgress.NextLogIndex)

	log.Printf("starting event syncing from block %d log %d", fromBlock, fromLogIndex)
	syncer := eventsyncer.New(chainobs.contracts.Client, finalityOffset, eventTypes, fromBlock, fromLogIndex)

	errorgroup, errorctx := errgroup.WithContext(ctx)
	errorgroup.Go(func() error {
		return syncer.Run(errorctx)
	})
	errorgroup.Go(func() error {
		for {
			eventSyncUpdate, err := syncer.Next(errorctx)
			if err != nil {
				return err
			}
			if err := chainobs.handleEventSyncUpdate(errorctx, eventSyncUpdate); err != nil {
				return err
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
		addrs, err := retryGetAddrs(ctx, chainobs.contracts.Collators, event.Index)
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
		db := commondb.New(tx)

		if eventSyncUpdate.Event != nil {
			if err := chainobs.handleEvent(ctx, db, eventSyncUpdate.Event); err != nil {
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
		if err := db.UpdateEventSyncProgress(ctx, commondb.UpdateEventSyncProgressParams{
			NextBlockNumber: int32(nextBlockNumber),
			NextLogIndex:    int32(nextLogIndex),
		}); err != nil {
			return errors.Wrap(err, "failed to update last synced event")
		}
		return nil
	})
}

func (chainobs *ChainObserver) handleEvent(
	ctx context.Context, db *commondb.Queries, event interface{},
) error {
	var err error
	switch event := event.(type) {
	case newKeyperConfig:
		err = chainobs.handleKeypersConfigsListNewConfigEvent(ctx, db, event)
	case newCollatorConfig:
		err = chainobs.handleCollatorConfigsListNewConfigEvent(ctx, db, event)
	default:
		log.Printf("ignoring unknown event %+v %T", event, event)
	}
	return err
}

func (chainobs *ChainObserver) handleKeypersConfigsListNewConfigEvent(
	ctx context.Context, db *commondb.Queries, event newKeyperConfig,
) error {
	log.Printf(
		"handling NewConfig event from keypers config contract in block %d (config index %d, activation block number %d)",
		event.Raw.BlockNumber, event.KeyperConfigIndex, event.ActivationBlockNumber,
	)
	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf(
			"activation block number %d from config contract would overflow int64",
			event.ActivationBlockNumber)
	}
	err := db.InsertKeyperSet(ctx, commondb.InsertKeyperSetParams{
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
	ctx context.Context, db *commondb.Queries, event newCollatorConfig,
) error {
	log.Printf(
		"handling NewConfig event from collator config contract in block %d (index %d, activation block number %d)",
		event.Raw.BlockNumber, event.Index, event.ActivationBlockNumber,
	)
	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf(
			"activation block number %d from config contract would overflow int64",
			event.ActivationBlockNumber,
		)
	}
	if len(event.addrs) > 1 {
		return errors.Errorf("got multiple collators from collator addrs set contract: %s", event.addrs)
	} else if len(event.addrs) == 1 {
		err := db.InsertChainCollator(ctx, commondb.InsertChainCollatorParams{
			ActivationBlockNumber: int64(event.ActivationBlockNumber),
			Collator:              shdb.EncodeAddress(event.addrs[0]),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to insert collator into db")
		}
	}
	return nil
}
