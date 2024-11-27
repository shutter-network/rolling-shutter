package gnosis

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"time"

	_ "github.com/lib/pq"
)

const (
	checkInterval = 30 * time.Second
)

type SyncMonitor struct {
	DBPool *pgxpool.Pool
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

	for {
		select {
		case <-time.After(checkInterval):
			record, err := db.GetTransactionSubmittedEventsSyncedUntil(ctx)
			if err != nil {
				log.Warn().Err(err).Msg("error fetching block number")
				continue
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
			return nil
		}
	}
}