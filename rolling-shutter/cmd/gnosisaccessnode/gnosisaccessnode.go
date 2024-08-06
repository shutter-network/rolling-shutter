package gnosisaccessnode

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/gnosisaccessnode"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		command.Usage(
			"Run an access node for the keyper network of Shutterized Gnosis Chain",
			`This command runs a node that only relays messages, but doesn't create any on
its own. It is intended to be a stable node to connect to to receive messages.`,
		),
		command.WithGenerateConfigSubcommand(),
		command.WithDumpConfigSubcommand(),
	)
	return builder.Command()
}

func main(config *gnosisaccessnode.Config) error {
	log.Info().
		Str("version", shversion.Version()).
		Msg("starting access node")

	accessNode := gnosisaccessnode.New(config)
	return service.RunWithSighandler(context.Background(), accessNode)
}
