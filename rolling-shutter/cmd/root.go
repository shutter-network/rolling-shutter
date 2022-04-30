// Package cmd implements the shuttermint subcommands
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/bootstrap"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/chain"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/collator"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/completion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/cryptocmd"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/mocknode"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/proxy"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/snapshot"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
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
	cmd.AddCommand(bootstrap.Cmd())
	cmd.AddCommand(chain.Cmd())
	cmd.AddCommand(collator.Cmd())
	cmd.AddCommand(completion.Cmd())
	cmd.AddCommand(keyper.Cmd())
	cmd.AddCommand(mocknode.Cmd())
	cmd.AddCommand(snapshot.Cmd())
	cmd.AddCommand(cryptocmd.Cmd())
	cmd.AddCommand(proxy.Cmd())
	return cmd
}
