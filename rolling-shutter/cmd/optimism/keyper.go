package optimism

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	keyper "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		command.Usage(
			"Run a Shutter optimism keyper node",
			`This command runs a keyper node. It will connect to both an Optimism and a
Shuttermint node which have to be started separately in advance.`,
		),
		command.WithGenerateConfigSubcommand(),
	)
	builder.AddInitDBCommand(initDB)
	return builder.Command()
}

func main(cfg *config.Config) error {
	log.Info().
		Str("version", shversion.Version()).
		Str("address", cfg.GetAddress().Hex()).
		Str("shuttermint", cfg.Shuttermint.ShuttermintURL).
		Msg("starting keyper")
	kpr, err := keyper.New(cfg)
	if err != nil {
		return err
	}
	return service.RunWithSighandler(context.Background(), kpr)
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
