package collator

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
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
	return db.InitDB(ctx, dbpool, database.Definition.Name(), database.Definition)
}

func main(cfg *config.Config) error {
	return service.RunWithSighandler(context.Background(), collator.New(cfg))
}
