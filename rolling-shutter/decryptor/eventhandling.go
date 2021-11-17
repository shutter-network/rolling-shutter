package decryptor

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shuttermint/contract"
	"github.com/shutter-network/shutter/shuttermint/contract/deployment"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley/eventsyncer"
)

const finalityOffset = 3

func (d *Decryptor) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		d.contracts.DecryptorsAppended,
		d.contracts.KeypersAppended,
	}

	eventSyncProgress, err := d.db.GetEventSyncProgress(ctx)
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
	syncer := eventsyncer.New(d.contracts.Client, finalityOffset, events, fromBlock, fromLogIndex)

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
			handler, err := d.newContractEventHandler(errorctx)
			if err != nil {
				return err
			}
			if err := handler.handleEventSyncUpdate(errorctx, eventSyncUpdate); err != nil {
				return err
			}
		}
	})
	return errorgroup.Wait()
}

// eventHandler isolates the parts of a decryptor that can be accessed when handling an event. For
// each new event, a new handler should be created and the handleEventSyncUpdate method be called
// once.
type eventHandler struct {
	tx        pgx.Tx
	db        *dcrdb.Queries
	contracts *deployment.Contracts
}

func (d *Decryptor) newContractEventHandler(ctx context.Context) (*eventHandler, error) {
	tx, err := d.dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	dbWithTx := d.db.WithTx(tx)
	return &eventHandler{
		tx:        tx,
		db:        dbWithTx,
		contracts: d.contracts,
	}, nil
}

// handleEventSyncUpdate handles events and advances the sync state, but rolls back any db updates
// on failure.
func (h *eventHandler) handleEventSyncUpdate(ctx context.Context, eventSyncUpdate eventsyncer.EventSyncUpdate) error {
	err := h.handleEventSyncUpdateDirty(ctx, eventSyncUpdate)
	if err != nil {
		errRollback := h.tx.Rollback(ctx)
		if errRollback != nil {
			log.Printf("error rolling back db transaction: %s", errRollback)
		}
		return err
	}
	err = h.tx.Commit(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to commit db tx after event was handled")
	}
	return nil
}

// handleEventSyncUpdateDirty handles events and advances the sync state. The db transaction will
// neither be committed nor rolled back at the end.
func (h *eventHandler) handleEventSyncUpdateDirty(ctx context.Context, eventSyncUpdate eventsyncer.EventSyncUpdate) error {
	switch event := eventSyncUpdate.Event.(type) {
	case contract.AddrsSeqAppended:
		switch event.Raw.Address {
		case h.contracts.KeypersAppended.Address:
			if err := h.handleKeypersAppendedEvent(ctx, event); err != nil {
				return err
			}
		case h.contracts.DecryptorsAppended.Address:
			if err := h.handleDecryptorsAppendedEvent(ctx, event); err != nil {
				return err
			}
		default:
			log.Printf("ignoring Appended event from unknown contract %s", event.Raw.Address)
		}
	case nil:
		// event is nil if no event is found for some time
	default:
		log.Printf("ignoring unknown event %+v %T", event, event)
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
	if err := h.db.UpdateEventSyncProgress(ctx, dcrdb.UpdateEventSyncProgressParams{
		NextBlockNumber: int32(nextBlockNumber),
		NextLogIndex:    int32(nextLogIndex),
	}); err != nil {
		return errors.Wrap(err, "failed to update last synced event")
	}
	return nil
}

func (h *eventHandler) handleKeypersAppendedEvent(ctx context.Context, event contract.AddrsSeqAppended) error {
	log.Println("handling Appended event from keypers")
	return nil
}

func (h *eventHandler) handleDecryptorsAppendedEvent(ctx context.Context, event contract.AddrsSeqAppended) error {
	log.Println("handling Appended event from decryptors")
	return nil
}
