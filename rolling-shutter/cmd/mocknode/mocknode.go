package mocknode

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocknode"
)

func Cmd() *cobra.Command {
	builder := command.Build(
		main,
		// TODO  long usage
		command.Usage(
			"Run a Shutter mock node",
			"",
		),
		command.WithGenerateConfigSubcommand(),
	)
	return builder.Command()
}

func main(cfg *mocknode.Config) error {
	mockNode, err := mocknode.New(cfg)
	if err != nil {
		return err
	}
	return service.RunWithSighandler(context.Background(), mockNode)
}
