package gnosis

import (
	"context"
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
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
	StartEon             uint64
	GenesisSlotTimestamp uint64
	SecondsPerSlot       uint64
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
		start = 0
	} else {
		start = uint64(syncedUntil.BlockNumber + 1)
	}

	log.Debug().
		Uint64("start-block", start).
		Uint64("end-block", header.Number.Uint64()).
		Msg("syncing sequencer contract")

	endBlock := header.Number.Uint64()
	opts := bind.FilterOpts{
		Start:   start,
		End:     &endBlock,
		Context: ctx,
	}
	it, err := s.Contract.SequencerFilterer.FilterTransactionSubmitted(&opts)
	if err != nil {
		return errors.Wrap(err, "failed to query transaction submitted events")
	}
	events := []*sequencerBindings.SequencerTransactionSubmitted{}
	for it.Next() {
		if it.Event.Eon < s.StartEon ||
			it.Event.Eon > math.MaxInt64 ||
			!it.Event.GasLimit.IsInt64() {
			log.Debug().
				Uint64("eon", it.Event.Eon).
				Uint64("block-number", it.Event.Raw.BlockNumber).
				Str("block-hash", it.Event.Raw.BlockHash.Hex()).
				Uint("tx-index", it.Event.Raw.TxIndex).
				Uint("log-index", it.Event.Raw.Index).
				Msg("ignoring transaction submitted event")
			continue
		}
		events = append(events, it.Event)
	}
	if it.Error() != nil {
		return errors.Wrap(it.Error(), "failed to iterate transaction submitted events")
	}
	if len(events) == 0 {
		log.Debug().
			Uint64("start-block", start).
			Uint64("end-block", endBlock).
			Msg("no transaction submitted events found")
	}

	return s.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		err = s.insertTransactionSubmittedEvents(ctx, tx, events)
		if err != nil {
			return err
		}

		newSyncedUntilBlock, err := medley.Uint64ToInt64Safe(endBlock)
		if err != nil {
			return err
		}
		slot := medley.BlockTimestampToSlot(header.Time, s.GenesisSlotTimestamp, s.SecondsPerSlot)
		err = queries.SetTransactionSubmittedEventsSyncedUntil(ctx, database.SetTransactionSubmittedEventsSyncedUntilParams{
			BlockNumber: newSyncedUntilBlock,
			BlockHash:   header.Hash().Bytes(),
			Slot:        int64(slot),
		})
		if err != nil {
			return err
		}
		return nil
	})
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
