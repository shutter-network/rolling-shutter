package db

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func ValidateDBVersion(ctx context.Context, dbpool *pgxpool.Pool, role string) error {
	err := dbpool.BeginFunc(
		ctx,
		func(tx pgx.Tx) error {
			return ValidateDatabaseVersion(ctx, tx, role)
		},
	)
	if err != nil {
		return errors.Wrap(err, "database is used for a different role already, preventing overwrite")
	}
	return nil
}

// Connect to the database `url` from within a runner.Start() method
// and create the pgxpool.Pool.
func Connect(ctx context.Context, runner service.Runner, url, version string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.Connect(ctx, url)
	if err != nil {
		return nil, err
	}
	runner.Defer(dbpool.Close)

	err = ValidateDBVersion(ctx, dbpool, version)
	if err != nil {
		return nil, err
	}
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("connected to database")
	return dbpool, nil
}
