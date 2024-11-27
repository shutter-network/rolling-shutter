package shutterservice

import (
	"bytes"
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
)

const AssumedReorgDepth = 10

type RegistrySyncer struct {
	Contract             common.Address //TODO: need to be changed to contract binding
	DBPool               *pgxpool.Pool
	ExecutionClient      *ethclient.Client
	SyncStartBlockNumber uint64
}

// getNumReorgedBlocks returns the number of blocks that have already been synced, but are no
// longer in the chain.
func getNumReorgedBlocks(syncedUntil *database.TransactionSubmittedEventsSyncedUntil, header *types.Header) int {
	shouldBeParent := header.Number.Int64() == syncedUntil.BlockNumber+1
	isParent := bytes.Equal(header.ParentHash.Bytes(), syncedUntil.BlockHash)
	isReorg := shouldBeParent && !isParent
	if !isReorg {
		return 0
	}
	// We don't know how deep the reorg is, so we make a conservative guess. Assuming higher depths
	// is safer because it means we resync a little bit more.
	depth := AssumedReorgDepth
	if syncedUntil.BlockNumber < int64(depth) {
		return int(syncedUntil.BlockNumber)
	}
	return depth
}

// resetSyncStatus clears the db from its recent history after a reorg of given depth.
func (s *RegistrySyncer) resetSyncStatus(ctx context.Context, numReorgedBlocks int) error {
	if numReorgedBlocks == 0 {
		return nil
	}
	return s.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		// queries := database.New(tx)

		// syncStatus, err := queries.GetTransactionSubmittedEventsSyncedUntil(ctx)
		// if err != nil {
		// 	return errors.Wrap(err, "failed to query sync status from db in order to reset it")
		// }
		// if syncStatus.BlockNumber < int64(numReorgedBlocks) {
		// 	return errors.Wrapf(err, "detected reorg deeper (%d) than blocks synced (%d)", syncStatus.BlockNumber, numReorgedBlocks)
		// }

		// deleteFromInclusive := syncStatus.BlockNumber - int64(numReorgedBlocks) + 1

		// err = queries.DeleteTransactionSubmittedEventsFromBlockNumber(ctx, deleteFromInclusive)
		// if err != nil {
		// 	return errors.Wrap(err, "failed to delete transaction submitted events from db")
		// }
		// Currently, we don't have enough information in the db to populate block hash and slot.
		// However, using default values here is fine since the syncer is expected to resync
		// immediately after this function call which will set the correct values. When we do proper
		// reorg handling, we should store the full block data of the previous blocks so that we can
		// avoid this.

		// newSyncedUntilBlockNumber := deleteFromInclusive - 1

		//TODO: need to change sync status to use registry event sync

		// err = queries.SetTransactionSubmittedEventsSyncedUntil(ctx, database.SetTransactionSubmittedEventsSyncedUntilParams{
		// 	BlockHash:   []byte{},
		// 	BlockNumber: newSyncedUntilBlockNumber,
		// 	Slot:        0,
		// })
		// if err != nil {
		// 	return errors.Wrap(err, "failed to reset transaction submitted event sync status in db")
		// }
		// log.Info().
		// 	Int("depth", numReorgedBlocks).
		// 	Int64("previous-synced-until", syncStatus.BlockNumber).
		// 	Int64("new-synced-until", newSyncedUntilBlockNumber).
		// 	Msg("sync status reset due to reorg")
		return nil
	})
}

// Sync fetches IdentityRegistered events from the registry contract and inserts them into the
// database. It starts at the end point of the previous call to sync (or 0 if it is the first call)
// and ends at the given block number.
func (s *RegistrySyncer) Sync(ctx context.Context, header *types.Header) error {
	//TODO: needs to be implemented
	return nil
}
