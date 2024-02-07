package gnosis

import (
	"context"
	"math"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
	Contract *sequencerBindings.Sequencer
	DBPool   *pgxpool.Pool
	StartEon uint64
}

func (s *SequencerSyncer) Sync(ctx context.Context, block uint64) error {
	queries := database.New(s.DBPool)
	syncedUntilBlock, err := queries.GetTransactionSubmittedEventsSyncedUntil(ctx)
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "failed to query transaction submitted events sync status")
	}
	var start uint64
	if err == pgx.ErrNoRows {
		start = 0
	} else {
		start = uint64(syncedUntilBlock) + 1
	}

	log.Debug().
		Uint64("start-block", start).
		Uint64("end-block", block).
		Msg("syncing sequencer contract")

	opts := bind.FilterOpts{
		Start:   start,
		End:     &block,
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
			Uint64("end-block", block).
			Msg("no transaction submitted events found")
		return nil
	}

	return s.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		queries := database.New(tx)
		for _, event := range events {
			_, err := queries.InsertTransactionSubmittedEvent(ctx, database.InsertTransactionSubmittedEventParams{
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
			log.Debug().
				Uint64("block", event.Raw.BlockNumber).
				Uint64("eon", event.Eon).
				Hex("identityPrefix", event.IdentityPrefix[:]).
				Hex("sender", event.Sender.Bytes()).
				Uint64("gasLimit", event.GasLimit.Uint64()).
				Msg("synced new transaction submitted event")
		}
		newSyncedUntilBlock, err := medley.Uint64ToInt64Safe(block)
		if err != nil {
			return err
		}
		return queries.SetTransactionSubmittedEventsSyncedUntil(ctx, newSyncedUntilBlock)
	})
}
