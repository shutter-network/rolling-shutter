package main

import (
	"github.com/spf13/cobra/doc"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/cmd"
)

func main() {
	err := doc.GenMarkdownTree(cmd.Command(), "./")
	if err != nil {
		panic(err)
	}
}
