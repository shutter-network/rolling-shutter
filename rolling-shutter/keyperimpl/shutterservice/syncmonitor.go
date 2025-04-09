package shutterservice

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	keyperDB "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
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
	keyperdb := keyperDB.New(s.DBPool)

	log.Debug().Msg("starting the sync monitor")

	for {
		select {
		case <-time.After(s.CheckInterval):
			if err := s.runCheck(ctx, db, keyperdb, &lastBlockNumber); err != nil {
				if errors.Is(err, ErrBlockNotIncreasing) {
					return err
				}
				log.Debug().Err(err).Msg("skipping sync check due to error")
			}
		case <-ctx.Done():
			log.Info().Msg("stopping syncMonitor due to context cancellation")
			return ctx.Err()
		}
	}
}

var ErrBlockNotIncreasing = errors.New("block number has not increased between checks")

func (s *SyncMonitor) runCheck(
	ctx context.Context,
	db *database.Queries,
	keyperdb *keyperDB.Queries,
	lastBlockNumber *int64,
) error {
	isRunning, err := s.isDKGRunning(ctx, keyperdb)
	if err != nil {
		return fmt.Errorf("syncMonitor | error in dkgIsRunning: %w", err)
	}
	if isRunning {
		log.Debug().Msg("dkg is running, skipping sync monitor checks")
		return nil
	}

	record, err := db.GetIdentityRegisteredEventsSyncedUntil(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn().Err(err).Msg("no rows found in table identity_registered_events_synced_until")
			return nil // This is not an error condition that should stop monitoring
		}
		return fmt.Errorf("error getting identity_registered_events_synced_until: %w", err)
	}

	currentBlockNumber := record.BlockNumber
	log.Debug().Int64("current-block-number", currentBlockNumber).Msg("current block number")

	if currentBlockNumber > *lastBlockNumber {
		*lastBlockNumber = currentBlockNumber
		return nil
	}

	log.Error().
		Int64("last-block-number", *lastBlockNumber).
		Int64("current-block-number", currentBlockNumber).
		Msg("block number has not increased between checks")
	return ErrBlockNotIncreasing
}

func (s *SyncMonitor) isDKGRunning(ctx context.Context, keyperdb *keyperDB.Queries) (bool, error) {
	// if latest eon is registered then EonStarted event has triggered, which means the dkg can start
	eons, err := keyperdb.GetAllEons(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		log.Error().
			Err(err).
			Msg("syncMonitor | error getting all eons")
		return false, err
	}

	if len(eons) == 0 {
		return false, nil
	}

	// if we get no rows in getting dkg result then dkg is not completed for that eon
	_, err = keyperdb.GetDKGResult(ctx, eons[len(eons)-1].Eon)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	} else if err != nil {
		log.Error().Err(err).Msg("syncMonitor | error getting dkg result")
		return false, err
	}
	return false, nil
}
