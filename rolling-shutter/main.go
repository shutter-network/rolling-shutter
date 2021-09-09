package main

//go:generate sqlc generate

import "github.com/shutter-network/shutter/shuttermint/cmd"

func main() {
	cmd.Execute()
}
