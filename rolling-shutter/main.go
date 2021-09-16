package main

//go:generate sqlc generate

import (
	"os"

	"github.com/shutter-network/shutter/shuttermint/cmd"
)

func main() {
	status := 0

	if err := cmd.Cmd().Execute(); err != nil {
		status = 1
	}

	os.Exit(status)
}
