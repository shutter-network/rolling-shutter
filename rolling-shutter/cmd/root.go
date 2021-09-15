// Package cmd implements the shuttermint subcommands
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/shutter-network/shutter/shuttermint/cmd/bootstrap"
	"github.com/shutter-network/shutter/shuttermint/cmd/keyper"
	"github.com/shutter-network/shutter/shuttermint/cmd/shversion"
	"github.com/shutter-network/shutter/shuttermint/medley"
)

var (
	cfgFile   string
	logformat string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
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
			return fmt.Errorf("bad log value, possible values: min, short, long, max")
		}

		log.SetFlags(flags)
		return nil
	},
	Run:          medley.ShowHelpAndExit,
	SilenceUsage: true,
}

// Execute the shuttermint root command and exit the program afterwards. This is called from main.
func Execute() {
	status := 0

	if err := rootCmd.Execute(); err != nil {
		status = 1
	}

	os.Exit(status)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logformat, "log", "long", "set log format, possible values:  min, short, long, max")
	rootCmd.AddCommand(chainCmd)
	rootCmd.AddCommand(keyper.Cmd())
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(txsearchCmd)
	rootCmd.AddCommand(bootstrap.Cmd())
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(decryptorCmd)
	rootCmd.AddCommand(completionCmd)
}
