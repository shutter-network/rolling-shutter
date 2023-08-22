package main

import (
	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/rootcmd"
)

//go:generate go run -C docs gendocs.go

func main() {
	rootcmd.Main(cmd.Command())
}
