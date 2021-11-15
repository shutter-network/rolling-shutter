package keyper

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/shutter-network/shutter/shuttermint/contract"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/medley/eventsyncer"
)

const finalityOffset = 3

func (kpr *keyper) handleContractEvents(ctx context.Context) error {
	events := []*eventsyncer.EventType{
		kpr.contracts.KeypersAppended,
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

func (kpr *keyper) handleEventSyncUpdate(ctx context.Context, eventSyncUpdate eventsyncer.EventSyncUpdate) error {
	switch event := eventSyncUpdate.Event.(type) {
	case contract.AddrsSeqAppended:
		switch event.Raw.Address {
		case kpr.contracts.KeypersAppended.Address:
			if err := kpr.handleKeypersAppendedEvent(ctx, event); err != nil {
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
	if err := kpr.db.UpdateEventSyncProgress(ctx, kprdb.UpdateEventSyncProgressParams{
		NextBlockNumber: int32(nextBlockNumber),
		NextLogIndex:    int32(nextLogIndex),
	}); err != nil {
		return errors.Wrap(err, "failed to update last synced event")
	}
	return nil
}

func (kpr *keyper) handleKeypersAppendedEvent(ctx context.Context, event contract.AddrsSeqAppended) error {
	log.Println("handling Appended event from keypers contract")
	return nil
}
