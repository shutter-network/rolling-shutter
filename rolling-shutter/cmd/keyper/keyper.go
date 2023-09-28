package keyper

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/kprdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	keyper "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/impl/rollup"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		command.Usage(
			"Run a Shutter keyper node",
			`This command runs a keyper node. It will connect to both an Ethereum and a
Shuttermint node which have to be started separately in advance.`,
		),
		command.WithGenerateConfigSubcommand(),
	)
	builder.AddInitDBCommand(initDB)
	return builder.Command()
}

func main(config *keyper.Config) error {
	log.Info().
		Str("version", shversion.Version()).
		Str("address", config.GetAddress().Hex()).
		Str("shuttermint", config.Shuttermint.ShuttermintURL).
		Msg("starting keyper")
	kpr, err := keyper.New(config)
	if err != nil {
		return err
	}
	return service.RunWithSighandler(context.Background(), kpr)
}

func initDB(config *keyper.Config) error {
	ctx := context.Background()

	dbpool, err := pgxpool.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer dbpool.Close()

	err = kprdb.ValidateKeyperDB(ctx, dbpool)
	if err == nil {
		shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database already exists")
		return nil
	} else if errors.Is(err, metadb.ErrSchemaMismatch) {
		return err
	}

	// initialize the db
	err = kprdb.InitDB(ctx, dbpool)
	if err != nil {
		return err
	}
	shdb.AddConnectionInfo(log.Info(), dbpool).Msg("database initialized")
	return nil
}
