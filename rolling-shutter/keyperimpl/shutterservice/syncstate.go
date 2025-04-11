package shutterservice

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
)

// ShutterServiceSyncState implements the BlockSyncState interface for the shutter service
type ShutterServiceSyncState struct {
	DBPool *pgxpool.Pool
}

// GetSyncedBlockNumber retrieves the current synced block number from identity events
func (s *ShutterServiceSyncState) GetSyncedBlockNumber(ctx context.Context) (int64, error) {
	db := database.New(s.DBPool)
	record, err := db.GetIdentityRegisteredEventsSyncedUntil(ctx)
	if err != nil {
		return 0, err
	}
	return record.BlockNumber, nil
}
