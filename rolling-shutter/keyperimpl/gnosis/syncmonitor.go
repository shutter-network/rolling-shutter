package gnosis

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type SyncMonitor struct {
	DBPool        *pgxpool.Pool
	CheckInterval time.Duration
}

func (s *SyncMonitor) Start(ctx context.Context, runner service.Runner) error {
	runner.Go(func() error {
		return s.runMonitor(ctx)
	})

	return nil
}

func (s *SyncMonitor) runMonitor(ctx context.Context) error {
	var lastBlockNumber int64
	db := database.New(s.DBPool)

	log.Debug().Msg("starting the sync monitor")

	for {
		select {
		case <-time.After(s.CheckInterval):
			record, err := db.GetTransactionSubmittedEventsSyncedUntil(ctx)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					log.Warn().Err(err).Msg("no rows found in table transaction_submitted_events_synced_until")
					continue
				}
				return errors.Wrap(err, "error getting transaction_submitted_events_synced_until")
			}

			currentBlockNumber := record.BlockNumber
			log.Debug().Int64("current-block-number", currentBlockNumber).Msg("current block number")

			if currentBlockNumber > lastBlockNumber {
				lastBlockNumber = currentBlockNumber
			} else {
				log.Error().
					Int64("last-block-number", lastBlockNumber).
					Int64("current-block-number", currentBlockNumber).
					Msg("block number has not increased between checks")
				return errors.New("block number has not increased between checks")
			}
		case <-ctx.Done():
			log.Info().Msg("stopping syncMonitor due to context cancellation")
			return ctx.Err()
		}
	}
}
