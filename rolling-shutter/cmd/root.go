// Package cmd implements the shuttermint subcommands
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/shutter-network/shutter/shuttermint/cmd/bootstrap"
	"github.com/shutter-network/shutter/shuttermint/cmd/chain"
	"github.com/shutter-network/shutter/shuttermint/cmd/completion"
	"github.com/shutter-network/shutter/shuttermint/cmd/decryptor"
	"github.com/shutter-network/shutter/shuttermint/cmd/keyper"
	"github.com/shutter-network/shutter/shuttermint/cmd/shversion"
	"github.com/shutter-network/shutter/shuttermint/medley"
)

var logformat string

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rolling-shutter",
		Short:   "A collection of commands to run and interact with Rolling Shutter nodes",
		Version: shversion.Version(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := medley.BindFlags(cmd, "ROLLING_SHUTTER")
			if err != nil {
				return err
			}
			var flags int

			switch logformat {
			case "min":
			case "short":
				flags = log.Lshortfile
			case "long":
				flags = log.LstdFlags | log.Lshortfile | log.Lmicroseconds
			case "max":
				flags = log.LstdFlags | log.Llongfile | log.Lmicroseconds
			default:
				return fmt.Errorf(
					"bad log value, possible values: min, short, long, max",
				)
			}

			log.SetFlags(flags)
			return nil
		},
		Run:          medley.ShowHelpAndExit,
		SilenceUsage: true,
	}
	cmd.PersistentFlags().StringVar(
		&logformat,
		"log",
		"long",
		"set log format, possible values:  min, short, long, max",
	)
	cmd.AddCommand(chain.Cmd())
	cmd.AddCommand(keyper.Cmd())
	cmd.AddCommand(bootstrap.Cmd())
	cmd.AddCommand(decryptor.Cmd())
	cmd.AddCommand(completion.Cmd())
	return cmd
}
