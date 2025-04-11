package gnosis

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
)

// GnosisSyncState implements the BlockSyncState interface for the gnosis keyper
type GnosisSyncState struct {
	DBPool *pgxpool.Pool
}

// GetSyncedBlockNumber retrieves the current synced block number from transaction submitted events
func (s *GnosisSyncState) GetSyncedBlockNumber(ctx context.Context) (int64, error) {
	db := database.New(s.DBPool)
	record, err := db.GetTransactionSubmittedEventsSyncedUntil(ctx)
	if err != nil {
		return 0, err
	}
	return record.BlockNumber, nil
}
