package collator

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		// TODO better usage descriptions
		command.Usage(
			"Run a collator node",
			"",
		),
		command.WithGenerateConfigSubcommand(),
	)
	builder.AddInitDBCommand(initDB)
	return builder.Command()
}

func initDB(cfg *config.Config) error {
	ctx := context.Background()

	dbpool, err := pgxpool.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()

	err = cltrdb.ValidateDB(ctx, dbpool)
	if err == nil {
		shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database already exists")
		return nil
	} else if errors.Is(err, metadb.ErrSchemaMismatch) {
		return err
	}

	// initialize the db
	err = cltrdb.InitDB(ctx, dbpool)
	if err != nil {
		return err
	}
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database initialized")
	return nil
}

func main(cfg *config.Config) error {
	return service.RunWithSighandler(context.Background(), collator.New(cfg))
}
