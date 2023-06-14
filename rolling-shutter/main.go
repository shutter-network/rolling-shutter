package main

import (
	"github.com/spf13/cobra"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/bootstrap"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/chain"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/collator"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/cryptocmd"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/mocknode"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/mocksequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/p2pnode"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/proxy"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/snapshot"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/snapshotkeyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/rootcmd"
)

func subcommands() []*cobra.Command {
	return []*cobra.Command{
		bootstrap.Cmd(),
		chain.Cmd(),
		collator.Cmd(),
		keyper.Cmd(),
		snapshotkeyper.Cmd(),
		mocknode.Cmd(),
		snapshot.Cmd(),
		cryptocmd.Cmd(),
		proxy.Cmd(),
		mocksequencer.Cmd(),
		p2pnode.Cmd(),
	}
}

func cmd() *cobra.Command {
	cmd := rootcmd.Cmd()
	cmd.Use = "rolling-shutter"
	cmd.Short = "A collection of commands to run and interact with Rolling Shutter nodes"
	cmd.AddCommand(subcommands()...)
	return cmd
}

func main() {
	rootcmd.Main(cmd.Command())
}
