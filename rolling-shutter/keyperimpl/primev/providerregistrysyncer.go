package primev

import (
	"bytes"
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	providerregistry "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/primev/abi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/primev/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
)

const (
	AssumedReorgDepth    = 10
	maxRequestBlockRange = 10_000
)

type ProviderRegistrySyncer struct {
	Contract             *providerregistry.Providerregistry
	DBPool               *pgxpool.Pool
	ExecutionClient      *ethclient.Client
	SyncStartBlockNumber uint64
}

// getNumReorgedBlocks returns the number of blocks that have already been synced, but are no
// longer in the chain.
func getNumReorgedBlocks(syncedUntil *database.ProviderRegistryEventsSyncedUntil, header *types.Header) int {
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
func (s *ProviderRegistrySyncer) resetSyncStatus(ctx context.Context, numReorgedBlocks int) error {
	if numReorgedBlocks == 0 {
		return nil
	}
	return s.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		queries := database.New(tx)

		syncStatus, err := queries.GetProviderRegistryEventsSyncedUntil(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to query sync status from db in order to reset it")
		}
		if syncStatus.BlockNumber < int64(numReorgedBlocks) {
			return errors.Wrapf(err, "detected reorg deeper (%d) than blocks synced (%d)", syncStatus.BlockNumber, numReorgedBlocks)
		}

		deleteFromInclusive := syncStatus.BlockNumber - int64(numReorgedBlocks) + 1

		err = queries.DeleteProviderRegistryEventsFromBlockNumber(ctx, deleteFromInclusive)
		if err != nil {
			return errors.Wrap(err, "failed to delete provider registered events from db")
		}
		// Currently, we don't have enough information in the db to populate block hash.
		// However, using default values here is fine since the syncer is expected to resync
		// immediately after this function call which will set the correct values. When we do proper
		// reorg handling, we should store the full block data of the previous blocks so that we can
		// avoid this.

		newSyncedUntilBlockNumber := deleteFromInclusive - 1

		err = queries.SetProviderRegistryEventsSyncedUntil(ctx, database.SetProviderRegistryEventsSyncedUntilParams{
			BlockHash:   []byte{},
			BlockNumber: newSyncedUntilBlockNumber,
		})
		if err != nil {
			return errors.Wrap(err, "failed to reset provider registered event sync status in db")
		}
		log.Info().
			Int("depth", numReorgedBlocks).
			Int64("previous-synced-until", syncStatus.BlockNumber).
			Int64("new-synced-until", newSyncedUntilBlockNumber).
			Msg("sync status reset due to reorg")
		return nil
	})
}

func (s *ProviderRegistrySyncer) handlePotentialReorg(ctx context.Context, header *types.Header) error {
	queries := database.New(s.DBPool)
	syncedUntil, err := queries.GetProviderRegistryEventsSyncedUntil(ctx)
	if err == pgx.ErrNoRows {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "failed to query registration events sync status")
	}

	numReorgedBlocks := getNumReorgedBlocks(&syncedUntil, header)
	if numReorgedBlocks > 0 {
		return s.resetSyncStatus(ctx, numReorgedBlocks)
	}
	return nil
}

// Sync fetches IdentityRegistered events from the registry contract and inserts them into the
// database. It starts at the end point of the previous call to sync (or 0 if it is the first call)
// and ends at the given block number.
func (s *ProviderRegistrySyncer) Sync(ctx context.Context, header *types.Header) error {
	if err := s.handlePotentialReorg(ctx, header); err != nil {
		return err
	}

	queries := database.New(s.DBPool)
	syncedUntil, err := queries.GetProviderRegistryEventsSyncedUntil(ctx)
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "failed to query provider registered events sync status")
	}
	var start uint64
	if err == pgx.ErrNoRows {
		start = s.SyncStartBlockNumber
	} else {
		start = uint64(syncedUntil.BlockNumber + 1) //nolint:gosec
	}
	endBlock := header.Number.Uint64()
	log.Debug().
		Uint64("start-block", start).
		Uint64("end-block", endBlock).
		Msg("syncing registry contract")
	syncRanges := medley.GetSyncRanges(start, endBlock, maxRequestBlockRange)
	for _, r := range syncRanges {
		err = s.syncRange(ctx, r[0], r[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ProviderRegistrySyncer) syncRange(
	ctx context.Context,
	start,
	end uint64,
) error {
	events, err := s.fetchEvents(ctx, start, end)
	if err != nil {
		return err
	}
	filteredEvents, blsKeys := s.filterEvents(ctx, events)

	header, err := s.ExecutionClient.HeaderByNumber(ctx, new(big.Int).SetUint64(end))
	if err != nil {
		return errors.Wrap(err, "failed to get execution block header by number")
	}
	err = s.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		err = s.insertProviderRegistryEvents(ctx, tx, filteredEvents, blsKeys)
		if err != nil {
			return err
		}
		return database.New(tx).SetProviderRegistryEventsSyncedUntil(ctx, database.SetProviderRegistryEventsSyncedUntilParams{
			BlockNumber: int64(end), //nolint:gosec
			BlockHash:   header.Hash().Bytes(),
		})
	})
	if err != nil {
		log.Warn().AnErr("error adding provider registered event into db", err)
	}
	log.Info().
		Uint64("start-block", start).
		Uint64("end-block", end).
		Int("num-inserted-events", len(filteredEvents)).
		Int("num-discarded-events", len(events)-len(filteredEvents)).
		Msg("synced provider registry contract")

	return nil
}

func (s *ProviderRegistrySyncer) fetchEvents(
	ctx context.Context,
	start,
	end uint64,
) ([]*providerregistry.ProviderregistryProviderRegistered, error) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     &end,
		Context: ctx,
	}

	// TODO: need to test if this fetches all provider registered events regardless of the provider address
	it, err := s.Contract.ProviderregistryFilterer.FilterProviderRegistered(&opts, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query provider registered events")
	}
	events := []*providerregistry.ProviderregistryProviderRegistered{}
	for it.Next() {
		events = append(events, it.Event)
	}
	if it.Error() != nil {
		return nil, errors.Wrap(it.Error(), "failed to iterate provider registered events")
	}
	return events, nil
}

func (s *ProviderRegistrySyncer) filterEvents(
	ctx context.Context,
	events []*providerregistry.ProviderregistryProviderRegistered,
) ([]*providerregistry.ProviderregistryProviderRegistered, [][][]byte) {
	filteredEvents := []*providerregistry.ProviderregistryProviderRegistered{}
	blsKeys := [][][]byte{}
	for _, event := range events {
		err := s.Contract.IsProviderValid(&bind.CallOpts{Context: ctx, BlockNumber: big.NewInt(int64(event.Raw.BlockNumber))}, event.Provider) //nolint:gosec
		if err != nil {
			log.Warn().
				Uint64("block-number", event.Raw.BlockNumber).
				Str("block-hash", event.Raw.BlockHash.Hex()).
				Uint("tx-index", event.Raw.TxIndex).
				Uint("log-index", event.Raw.Index).
				Str("provider", event.Provider.Hex()).
				Msg("ignoring provider registered event with invalid provider")
			continue
		}

		blsKey, err := s.Contract.GetBLSKeys(&bind.CallOpts{Context: ctx, BlockNumber: big.NewInt(int64(event.Raw.BlockNumber))}, event.Provider) //nolint:gosec
		if err != nil {
			log.Warn().
				Uint64("block-number", event.Raw.BlockNumber).
				Str("block-hash", event.Raw.BlockHash.Hex()).
				Uint("tx-index", event.Raw.TxIndex).
				Uint("log-index", event.Raw.Index).
				Str("provider", event.Provider.Hex()).
				Msg("ignoring provider registered event with invalid provider")
			continue
		}

		filteredEvents = append(filteredEvents, event)
		blsKeys = append(blsKeys, blsKey)
	}
	return filteredEvents, blsKeys
}

// insertProviderRegistryEvents inserts the given events into the database.
func (s *ProviderRegistrySyncer) insertProviderRegistryEvents(
	ctx context.Context,
	tx pgx.Tx,
	events []*providerregistry.ProviderregistryProviderRegistered,
	blsKeys [][][]byte,
) error {
	queries := database.New(tx)
	for i, event := range events {
		_, err := queries.InsertProviderRegistryEvent(ctx, database.InsertProviderRegistryEventParams{
			BlockNumber:     int64(event.Raw.BlockNumber),
			BlockHash:       event.Raw.BlockHash.Bytes(),
			TxIndex:         int64(event.Raw.TxIndex), //nolint:gosec
			LogIndex:        int64(event.Raw.Index),   //nolint:gosec
			ProviderAddress: event.Provider.Hex(),
			BlsKeys:         blsKeys[i],
		})
		if err != nil {
			return errors.Wrap(err, "failed to insert provider registered event into db")
		}
		log.Debug().
			Uint64("block", event.Raw.BlockNumber).
			Str("provider", event.Provider.Hex()).
			Msg("synced new provider registered event")
	}
	return nil
}
