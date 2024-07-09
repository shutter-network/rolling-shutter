package roguenode

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/roguenode"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		command.Usage(
			"Run a rogue node that sends malicious messages",
			`This command runs a node that sends malicious messages to its
peers, testing their resilience.`,
		),
		command.WithGenerateConfigSubcommand(),
		command.WithDumpConfigSubcommand(),
	)
	return builder.Command()
}

func main(config *roguenode.Config) error {
	log.Info().
		Str("version", shversion.Version()).
		Msg("starting rogue node")

	rogueNode := roguenode.New(config)
	return service.RunWithSighandler(context.Background(), rogueNode)
}
