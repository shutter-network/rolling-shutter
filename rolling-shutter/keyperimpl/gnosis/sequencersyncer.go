package gnosis

import (
	"context"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// SequencerSyncer inserts transaction submitted events from the sequencer contract into the database.
type SequencerSyncer struct {
	Contract             *sequencerBindings.Sequencer
	DBPool               *pgxpool.Pool
	ExecutionClient      *ethclient.Client
	GenesisSlotTimestamp uint64
	SecondsPerSlot       uint64
	SyncStartBlockNumber uint64
}

// Sync fetches transaction submitted events from the sequencer contract and inserts them into the
// database. It starts at the end point of the previous call to sync (or 0 if it is the first call)
// and ends at the given block number.
func (s *SequencerSyncer) Sync(ctx context.Context, header *types.Header) error {
	queries := database.New(s.DBPool)
	syncedUntil, err := queries.GetTransactionSubmittedEventsSyncedUntil(ctx)
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "failed to query transaction submitted events sync status")
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
		Msg("syncing sequencer contract")

	syncRanges := medley.GetSyncRanges(start, endBlock, maxRequestBlockRange)
	for _, r := range syncRanges {
		err = s.syncRange(ctx, r[0], r[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SequencerSyncer) syncRange(
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
		err = s.insertTransactionSubmittedEvents(ctx, tx, filteredEvents)
		if err != nil {
			return err
		}

		slot := medley.BlockTimestampToSlot(header.Time, s.GenesisSlotTimestamp, s.SecondsPerSlot)
		return database.New(tx).SetTransactionSubmittedEventsSyncedUntil(ctx, database.SetTransactionSubmittedEventsSyncedUntilParams{
			BlockNumber: int64(end),
			BlockHash:   header.Hash().Bytes(),
			Slot:        int64(slot),
		})
	})
	log.Info().
		Uint64("start-block", start).
		Uint64("end-block", end).
		Int("num-inserted-events", len(filteredEvents)).
		Int("num-discarded-events", len(events)-len(filteredEvents)).
		Msg("synced sequencer contract")
	metricsTxSubmittedEventsSyncedUntil.Set(float64(end))
	return nil
}

func (s *SequencerSyncer) fetchEvents(
	ctx context.Context,
	start,
	end uint64,
) ([]*sequencerBindings.SequencerTransactionSubmitted, error) {
	opts := bind.FilterOpts{
		Start:   start,
		End:     &end,
		Context: ctx,
	}
	it, err := s.Contract.SequencerFilterer.FilterTransactionSubmitted(&opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query transaction submitted events")
	}
	events := []*sequencerBindings.SequencerTransactionSubmitted{}
	for it.Next() {
		events = append(events, it.Event)
	}
	if it.Error() != nil {
		return nil, errors.Wrap(it.Error(), "failed to iterate transaction submitted events")
	}
	return events, nil
}

func (s *SequencerSyncer) filterEvents(
	events []*sequencerBindings.SequencerTransactionSubmitted,
) []*sequencerBindings.SequencerTransactionSubmitted {
	filteredEvents := []*sequencerBindings.SequencerTransactionSubmitted{}
	for _, event := range events {
		if event.Eon > math.MaxInt64 ||
			!event.GasLimit.IsInt64() {
			log.Debug().
				Uint64("eon", event.Eon).
				Uint64("block-number", event.Raw.BlockNumber).
				Str("block-hash", event.Raw.BlockHash.Hex()).
				Uint("tx-index", event.Raw.TxIndex).
				Uint("log-index", event.Raw.Index).
				Msg("ignoring transaction submitted event with high eon")
			continue
		}
		filteredEvents = append(filteredEvents, event)
	}
	return filteredEvents
}

// insertTransactionSubmittedEvents inserts the given events into the database and updates the
// transaction submitted event number accordingly.
func (s *SequencerSyncer) insertTransactionSubmittedEvents(
	ctx context.Context,
	tx pgx.Tx,
	events []*sequencerBindings.SequencerTransactionSubmitted,
) error {
	queries := database.New(tx)
	nextEventIndices := make(map[uint64]int64)
	for _, event := range events {
		nextEventIndex, ok := nextEventIndices[event.Eon]
		if !ok {
			nextEventIndexFromDB, err := queries.GetTransactionSubmittedEventCount(ctx, int64(event.Eon))
			if err == pgx.ErrNoRows {
				nextEventIndexFromDB = 0
			} else if err != nil {
				return errors.Wrapf(err, "failed to query count of transaction submitted events for eon %d", event.Eon)
			}
			nextEventIndices[event.Eon] = nextEventIndexFromDB
			nextEventIndex = nextEventIndexFromDB
		}

		_, err := queries.InsertTransactionSubmittedEvent(ctx, database.InsertTransactionSubmittedEventParams{
			Index:          nextEventIndex,
			BlockNumber:    int64(event.Raw.BlockNumber),
			BlockHash:      event.Raw.BlockHash[:],
			TxIndex:        int64(event.Raw.TxIndex),
			LogIndex:       int64(event.Raw.Index),
			Eon:            int64(event.Eon),
			IdentityPrefix: event.IdentityPrefix[:],
			Sender:         shdb.EncodeAddress(event.Sender),
			GasLimit:       event.GasLimit.Int64(),
		})
		if err != nil {
			return errors.Wrap(err, "failed to insert transaction submitted event into db")
		}
		metricsLatestTxSubmittedEventIndex.WithLabelValues(string(event.Eon)).Set(float64(nextEventIndex))
		nextEventIndices[event.Eon]++
		log.Debug().
			Int64("index", nextEventIndex).
			Uint64("block", event.Raw.BlockNumber).
			Uint64("eon", event.Eon).
			Hex("identityPrefix", event.IdentityPrefix[:]).
			Hex("sender", event.Sender.Bytes()).
			Uint64("gasLimit", event.GasLimit.Uint64()).
			Msg("synced new transaction submitted event")
	}
	for eon, nextEventIndex := range nextEventIndices {
		err := queries.SetTransactionSubmittedEventCount(ctx, database.SetTransactionSubmittedEventCountParams{
			Eon:        int64(eon),
			EventCount: nextEventIndex,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
