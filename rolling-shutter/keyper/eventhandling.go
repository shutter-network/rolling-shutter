package keyper

import (
	"context"
	"log"
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shuttermint/contract"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/medley/eventsyncer"
	"github.com/shutter-network/shutter/shuttermint/shdb"
)

const finalityOffset = 3

func (kpr *keyper) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		kpr.contracts.KeypersConfigsListNewConfig,
		kpr.contracts.CollatorConfigsListNewConfig,
	}

	eventSyncProgress, err := kpr.db.GetEventSyncProgress(ctx)
	var fromBlock uint64
	var fromLogIndex uint64
	if err == pgx.ErrNoRows {
		fromBlock = 0
		fromLogIndex = 0
	} else if err == nil {
		fromBlock = uint64(eventSyncProgress.NextBlockNumber)
		fromLogIndex = uint64(eventSyncProgress.NextLogIndex)
	} else {
		return errors.Wrap(err, "failed to get last synced event from db")
	}

	log.Printf("starting event syncing from block %d log %d", fromBlock, fromLogIndex)
	syncer := eventsyncer.New(kpr.contracts.Client, finalityOffset, events, fromBlock, fromLogIndex)

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
			if err := kpr.handleEventSyncUpdate(errorctx, eventSyncUpdate); err != nil {
				return err
			}
		}
	})
	return errorgroup.Wait()
}

type eventHandler struct {
	tx        pgx.Tx
	db        *kprdb.Queries
	contracts *deployment.Contracts
}

// handleEventSyncUpdate handles events and advances the sync state, but rolls back any db updates
// on failure.
func (kpr *keyper) handleEventSyncUpdate(ctx context.Context, eventSyncUpdate eventsyncer.EventSyncUpdate) (rErr error) {
	// Create a db tx that we either commit or rollback at the end of the function, depending on if
	// an error is returned or not.
	tx, err := kpr.dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errors.Wrapf(err, "error committing db transaction")
	}
	defer func() {
		if rErr == nil {
			rErr = tx.Commit(ctx)
			return
		}
		if err := tx.Rollback(ctx); err != nil {
			log.Printf("error rolling back db transaction after failed event handling: %s", err)
		}
	}()
	dbWithTx := kpr.db.WithTx(tx)

	if eventSyncUpdate.Event != nil {
		handler := &eventHandler{
			tx:        tx,
			db:        dbWithTx,
			contracts: kpr.contracts,
		}
		if err := handler.handleEvent(ctx, eventSyncUpdate.Event); err != nil {
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
	if err := dbWithTx.UpdateEventSyncProgress(ctx, kprdb.UpdateEventSyncProgressParams{
		NextBlockNumber: int32(nextBlockNumber),
		NextLogIndex:    int32(nextLogIndex),
	}); err != nil {
		return errors.Wrap(err, "failed to update last synced event")
	}
	return nil
}

// handleEventSyncUpdateDirty handles events and advances the sync state. The db transaction will
// neither be committed nor rolled back at the end.
func (h *eventHandler) handleEvent(ctx context.Context, event interface{}) error {
	var err error
	switch event := event.(type) {
	case contract.KeypersConfigsListNewConfig:
		err = h.handleKeypersConfigsListNewConfigEvent(ctx, event)
	case contract.CollatorConfigsListNewConfig:
		err = h.handleCollatorConfigsListNewConfigEvent(ctx, event)
	default:
		log.Printf("ignoring unknown event %+v %T", event, event)
	}
	return err
}

func (h *eventHandler) handleKeypersConfigsListNewConfigEvent(ctx context.Context, event contract.KeypersConfigsListNewConfig) error {
	log.Printf(
		"handling NewConfig event from keypers config contract in block %d (index %d, activation block number %d)",
		event.Raw.BlockNumber, event.Index, event.ActivationBlockNumber,
	)
	callOpts := &bind.CallOpts{
		Pending: false,
		// We call for the current height instead of the height at which the event was emitted,
		// because the sets cannot change retroactively and we won't need an archive node.
		BlockNumber: nil,
		Context:     ctx,
	}
	addrsUntyped, err := medley.Retry(ctx, func() (interface{}, error) {
		return h.contracts.Keypers.GetAddrs(callOpts, event.Index)
	})
	if err != nil {
		return errors.Wrapf(err, "failed to query keyper addrs set from contract")
	}
	addrs := addrsUntyped.([]common.Address)

	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf("activation block number %d from config contract would overflow int64", event.ActivationBlockNumber)
	}
	err = h.db.InsertChainKeyperSet(ctx, kprdb.InsertChainKeyperSetParams{
		ActivationBlockNumber: int64(event.ActivationBlockNumber),
		Keypers:               shdb.EncodeAddresses(addrs),
		Threshold:             int32(event.Threshold),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to insert keyper set into db")
	}
	return nil
}

func (h *eventHandler) handleCollatorConfigsListNewConfigEvent(ctx context.Context, event contract.CollatorConfigsListNewConfig) error {
	log.Printf(
		"handling NewConfig event from collator config contract in block %d (index %d, activation block number %d)",
		event.Raw.BlockNumber, event.Index, event.ActivationBlockNumber,
	)
	callOpts := &bind.CallOpts{
		Pending: false,
		// We call for the current height instead of the height at which the event was emitted,
		// because the sets cannot change retroactively and we won't need an archive node.
		BlockNumber: nil,
		Context:     ctx,
	}
	addrsUntyped, err := medley.Retry(ctx, func() (interface{}, error) {
		return h.contracts.Collators.GetAddrs(callOpts, event.Index)
	})
	if err != nil {
		return errors.Wrapf(err, "failed to query keyper addrs set from contract")
	}
	addrs := addrsUntyped.([]common.Address)

	if event.ActivationBlockNumber > math.MaxInt64 {
		return errors.Errorf("activation block number %d from config contract would overflow int64", event.ActivationBlockNumber)
	}
	if len(addrs) > 1 {
		return errors.Errorf("got multiple collators from collator addrs set contract: %s", addrs)
	} else if len(addrs) == 1 {
		err = h.db.InsertChainCollator(ctx, kprdb.InsertChainCollatorParams{
			ActivationBlockNumber: int64(event.ActivationBlockNumber),
			Collator:              shdb.EncodeAddress(addrs[0]),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to insert collator into db")
		}
	}
	return nil
}
