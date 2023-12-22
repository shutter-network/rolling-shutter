package chainobserver

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	syncdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/sync"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
)

var ErrDBUpdateFail = errors.New("failed to update last synced event")

// handleEventSyncUpdate handles events and advances the sync state, but rolls back any db updates
// on failure.
func (c *ChainObserver) handleEventSyncUpdate(
	ctx context.Context, eventSyncUpdate eventsyncer.EventSyncUpdate,
) error {
	return c.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		if eventSyncUpdate.Event != nil {
			if err := c.handleEvent(ctx, tx, eventSyncUpdate.Event); err != nil {
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
		db := syncdb.New(tx)
		if err := db.UpdateEventSyncProgress(ctx, syncdb.UpdateEventSyncProgressParams{
			NextBlockNumber: int32(nextBlockNumber),
			NextLogIndex:    int32(nextLogIndex),
		}); err != nil {
			return errors.Wrap(err, ErrDBUpdateFail.Error())
		}
		return nil
	})
}
