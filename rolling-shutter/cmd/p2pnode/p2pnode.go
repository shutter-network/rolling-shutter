package p2pnode

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pnode"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		// TODO  long usage
		command.Usage(
			"Run a Shutter p2p bootstrap node",
			"",
		),
		command.WithGenerateConfigSubcommand(),
		command.WithDumpConfigSubcommand(),
	)
	return builder.Command()
}

func main(config *p2pnode.Config) error {
	log.Info().
		Str("version", shversion.Version()).
		Msg("starting p2pnode")
	p2pService, err := p2pnode.New(config)
	if err != nil {
		return errors.Wrapf(err, "failed to instantiate p2pnode")
	}
	return service.RunWithSighandler(context.Background(), p2pService)
}
