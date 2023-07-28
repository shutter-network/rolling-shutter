package main

import (
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/rootcmd"
)

func main() {
	rootcmd.Main(cmd.Command())
}
