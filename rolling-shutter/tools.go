//go:build tools
// +build tools

// Package tools is used to declare and track tool dependencies.  See
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module for
// further information.
package tools

import (
	_ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
	_ "github.com/ethereum/go-ethereum/cmd/abigen"
	_ "github.com/kyleconroy/sqlc/cmd/sqlc"
	_ "golang.org/x/tools/cmd/stringer"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
