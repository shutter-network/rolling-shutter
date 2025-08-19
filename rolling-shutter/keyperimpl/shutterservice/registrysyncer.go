package shutterservice

import (
	"bytes"
	"context"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	registryBindings "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/help"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

const (
	AssumedReorgDepth    = 10
	maxRequestBlockRange = 10_000
)

type RegistrySyncer struct {
	Contract             *registryBindings.ShutterRegistry
	DBPool               *pgxpool.Pool
	ExecutionClient      *ethclient.Client
	SyncStartBlockNumber uint64
}

// getNumReorgedBlocks returns the number of blocks that have already been synced, but are no
// longer in the chain.
func getNumReorgedBlocks(syncedUntil *database.IdentityRegisteredEventsSyncedUntil, header *types.Header) int {
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
		queries := database.New(tx)

		syncStatus, err := queries.GetIdentityRegisteredEventsSyncedUntil(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to query sync status from db in order to reset it")
		}
		if syncStatus.BlockNumber < int64(numReorgedBlocks) {
			return errors.Wrapf(err, "detected reorg deeper (%d) than blocks synced (%d)", syncStatus.BlockNumber, numReorgedBlocks)
		}

		deleteFromInclusive := syncStatus.BlockNumber - int64(numReorgedBlocks) + 1

		err = queries.DeleteIdentityRegisteredEventsFromBlockNumber(ctx, deleteFromInclusive)
		if err != nil {
			return errors.Wrap(err, "failed to delete identity registered events from db")
		}
		// Currently, we don't have enough information in the db to populate block hash.
		// However, using default values here is fine since the syncer is expected to resync
		// immediately after this function call which will set the correct values. When we do proper
		// reorg handling, we should store the full block data of the previous blocks so that we can
		// avoid this.

		newSyncedUntilBlockNumber := deleteFromInclusive - 1

		err = queries.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
			BlockHash:   []byte{},
			BlockNumber: newSyncedUntilBlockNumber,
		})
		if err != nil {
			return errors.Wrap(err, "failed to reset identity registered event sync status in db")
		}
		log.Info().
			Int("depth", numReorgedBlocks).
			Int64("previous-synced-until", syncStatus.BlockNumber).
			Int64("new-synced-until", newSyncedUntilBlockNumber).
			Msg("sync status reset due to reorg")
		return nil
	})
}

func (s *RegistrySyncer) handlePotentialReorg(ctx context.Context, header *types.Header) error {
	queries := database.New(s.DBPool)
	syncedUntil, err := queries.GetIdentityRegisteredEventsSyncedUntil(ctx)
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
func (s *RegistrySyncer) Sync(ctx context.Context, header *types.Header) error {
	if err := s.handlePotentialReorg(ctx, header); err != nil {
		return err
	}

	queries := database.New(s.DBPool)
	syncedUntil, err := queries.GetIdentityRegisteredEventsSyncedUntil(ctx)
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "failed to query identity registered events sync status")
	}
	var start uint64
	if err == pgx.ErrNoRows {
		start = s.SyncStartBlockNumber
	} else {
		start = uint64(syncedUntil.BlockNumber + 1)
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

func (s *RegistrySyncer) syncRange(
	ctx context.Context,
	start,
	end uint64,
) error {
	events, err := s.fetchEvents(ctx, start, end)
	if err != nil {
		return err
	}
	filteredEvents := s.filterEvents(events)

	header, err := s.ExecutionClient.HeaderByNumber(ctx, new(big.Int).SetUint64(end))
	if err != nil {
		return errors.Wrap(err, "failed to get execution block header by number")
	}
	err = s.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		err = s.insertIdentityRegisteredEvents(ctx, tx, filteredEvents)
		if err != nil {
			return err
		}
		return database.New(tx).SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
			BlockNumber: int64(end),
			BlockHash:   header.Hash().Bytes(),
		})
	})
	if err != nil {
		log.Warn().AnErr("error adding identity registered event into db", err)
	}
	log.Info().
		Uint64("start-block", start).
		Uint64("end-block", end).
		Int("num-inserted-events", len(filteredEvents)).
		Int("num-discarded-events", len(events)-len(filteredEvents)).
		Msg("synced registry contract")

	metricsRegistryEventsSyncedUntil.Set(float64(end))
	return nil
}

func (s *RegistrySyncer) fetchEvents(
	ctx context.Context,
	start,
	end uint64,
) ([]*registryBindings.ShutterRegistryEventTriggerRegistered, error) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     &end,
		Context: ctx,
	}
	it, err := s.Contract.ShutterRegistryFilterer.FilterEventTriggerRegistered(&opts, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query identity registered events")
	}
	events := []*registryBindings.ShutterRegistryEventTriggerRegistered{}
	for it.Next() {
		events = append(events, it.Event)
	}
	if it.Error() != nil {
		return nil, errors.Wrap(it.Error(), "failed to iterate identity registered events")
	}
	return events, nil
}

func (s *RegistrySyncer) filterEvents(
	events []*registryBindings.ShutterRegistryEventTriggerRegistered,
) []*registryBindings.ShutterRegistryEventTriggerRegistered {
	filteredEvents := []*registryBindings.ShutterRegistryEventTriggerRegistered{}
	for _, event := range events {
		if event.Eon > math.MaxInt64 {
			log.Debug().
				Uint64("eon", event.Eon).
				Uint64("block-number", event.Raw.BlockNumber).
				Str("block-hash", event.Raw.BlockHash.Hex()).
				Uint("tx-index", event.Raw.TxIndex).
				Uint("log-index", event.Raw.Index).
				Msg("ignoring identity registered event with high eon")
			continue
		}
		filteredEvents = append(filteredEvents, event)
	}
	return filteredEvents
}

// insertIdentityRegisteredEvents inserts the given events into the database.
func (s *RegistrySyncer) insertIdentityRegisteredEvents(
	ctx context.Context,
	tx pgx.Tx,
	events []*registryBindings.ShutterRegistryEventTriggerRegistered,
) error {
	queries := database.New(tx)
	for _, event := range events {
		identity := computeIdentity(event)
		var def []byte
		for i := range event.TriggerDefinition {
			def = append(def, event.TriggerDefinition[i]...)
		}
		parsed := EventTriggerDefinition{}
		err := parsed.UnmarshalBytes(event.TriggerDefinition)
		if err != nil {
			return errors.Wrap(err, "could not parse event definition from event")
		}
		_, err = queries.InsertEventTriggerRegisteredEvent(ctx, database.InsertEventTriggerRegisteredEventParams{
			BlockNumber:    int64(event.Raw.BlockNumber),
			BlockHash:      event.Raw.BlockHash[:],
			TxIndex:        int64(event.Raw.TxIndex),
			LogIndex:       int64(event.Raw.Index),
			Eon:            int64(event.Eon),
			IdentityPrefix: event.IdentityPrefix[:],
			Sender:         shdb.EncodeAddress(event.Sender),
			Definition:     string(def[:]),
			Ttl:            event.Ttl.Int64(),
			Identity:       identity,
		})
		if err != nil {
			return errors.Wrap(err, "failed to insert identity registered event into db")
		}
		log.Debug().
			Uint64("block", event.Raw.BlockNumber).
			Uint64("eon", event.Eon).
			Hex("identityPrefix", event.IdentityPrefix[:]).
			Hex("sender", event.Sender.Bytes()).
			Int64("ttl", event.Ttl.Int64()).
			Msg("synced new identity registered event")
	}
	return nil
}

func computeIdentity(event *registryBindings.ShutterRegistryEventTriggerRegistered) []byte {
	// TODO: may need to change this if we want to create identity other way
	var buf bytes.Buffer
	buf.Write(event.IdentityPrefix[:])
	buf.Write(event.Sender.Bytes())
	return crypto.Keccak256(buf.Bytes())
}
