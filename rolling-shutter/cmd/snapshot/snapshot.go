package snapshot

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/snpdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/snapshot"
)

var (
	cfgFile    string
	outputFile string
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		command.Usage(
			"Run the Snapshot Hub communication module",
			// TODO long usage description
			"",
		),
		command.WithGenerateConfigSubcommand(),
	)
	builder.AddInitDBCommand(initDB)
	return builder.Command()
}

func initDB(config *snapshot.Config) error {
	ctx := context.Background()

	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()

	err = snpdb.ValidateSnapshotDB(ctx, dbpool)
	if err == nil {
		shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database already exists")
		return nil
	} else if errors.Is(err, metadb.ErrSchemaMismatch) {
		return err
	}

	// initialize the db
	err = snpdb.InitDB(ctx, dbpool)
	if err != nil {
		return err
	}
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database initialized")

	return nil
}

func main(config *snapshot.Config) error {
	snp, err := snapshot.New(config)
	if err != nil {
		return err
	}
	return service.RunWithSighandler(
		context.Background(),
		snp,
	)
}
