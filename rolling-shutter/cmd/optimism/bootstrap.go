package optimism

import (
	"context"

	"github.com/spf13/cobra"

	boot "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/bootstrap"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
)

// TODO: use this to replace the old bootstrap command.
// First writing the keyperset and then bootstrapping allows
// to support different contracts etc.
func OPBootstrapCmd() *cobra.Command {
	builder := command.Build(
		bootstrap,
		command.Usage(
			"Bootstrap validator utility functions for a shuttermint chain",
			``,
		),
		command.WithGenerateConfigSubcommand(),
	)

	bootstrapCmd := &cobra.Command{
		Use:   "fetch-keyperset",
		Short: "fetch-keyperset",
		Args:  cobra.NoArgs,
		RunE:  builder.WrapFuncParseConfig(keyperSet),
	}
	builder.Command().AddCommand(bootstrapCmd)
	return builder.Command()
}

func keyperSet(cfg *boot.Config) error {
	ctx := context.Background()
	return boot.GetKeyperSet(ctx, cfg)
}

func bootstrap(cfg *boot.Config) error {
	return boot.BootstrapValidators(cfg)
}
