package cmd

import (
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/bootstrap"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/chain"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/collator"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/cryptocmd"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/gnosisaccessnode"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/gnosiskeyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/mocknode"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/mocksequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/optimism"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/p2pnode"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/proxy"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/rollupkeyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/snapshot"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/snapshotkeyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/rootcmd"
)

func Subcommands() []*cobra.Command {
	return []*cobra.Command{
		bootstrap.Cmd(),
		chain.Cmd(),
		collator.Cmd(),
		optimism.Cmd(),
		optimism.OPBootstrapCmd(),
		rollupkeyper.Cmd(),
		mocknode.Cmd(),
		snapshot.Cmd(),
		snapshotkeyper.Cmd(),
		gnosiskeyper.Cmd(),
		gnosisaccessnode.Cmd(),
		cryptocmd.Cmd(),
		proxy.Cmd(),
		mocksequencer.Cmd(),
		p2pnode.Cmd(),
	}
}

func Command() *cobra.Command {
	cmd := rootcmd.Cmd()
	cmd.Use = "rolling-shutter"
	cmd.Short = "A collection of commands to run and interact with Rolling Shutter nodes"
	cmd.AddCommand(Subcommands()...)
	return cmd
}
