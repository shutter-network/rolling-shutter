package gnosiskeyper

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/gnosiskeyperwatcher"
	keyper "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis"
	keyperconfig "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		command.Usage(
			"Run a Shutter keyper for Gnosis Chain",
			`This command runs a keyper node. It will connect to both a Gnosis and a
Shuttermint node which have to be started separately in advance.`,
		),
		command.WithGenerateConfigSubcommand(),
		command.WithDumpConfigSubcommand(),
	)
	builder.AddInitDBCommand(initDB)
	builder.AddFunctionSubcommand(
		watch,
		"watch",
		"Watch the keypers doing their work and log the generated decryption keys.",
		cobra.NoArgs,
	)
	return builder.Command()
}

func main(config *keyperconfig.Config) error {
	log.Info().
		Str("version", shversion.Version()).
		Str("address", config.GetAddress().Hex()).
		Str("shuttermint", config.Shuttermint.ShuttermintURL).
		Msg("starting gnosis keyper")

	kpr := keyper.New(config)
	return service.RunWithSighandler(context.Background(), kpr)
}

func initDB(cfg *keyperconfig.Config) error {
	ctx := context.Background()
	dbpool, err := pgxpool.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()
	return db.InitDB(ctx, dbpool, database.Definition.Name(), database.Definition)
}

func watch(cfg *keyperconfig.Config) error {
	log.Info().Msg("starting monitor")
	return service.RunWithSighandler(context.Background(), gnosiskeyperwatcher.New(cfg))
}
