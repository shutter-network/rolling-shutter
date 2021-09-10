# shuttermint

## Installation

Make sure you have at least go version 1.16 installed. Make sure `PATH` contains
`$GOPATH/bin`. If you didn't set `GOPATH`, it defaults to `${HOME}/go`.

Run `make` or `make build` to build the executables. The executables are build
in the `bin` directory.

Run `make install-tools` to install additional tools for linting and code
generation (sqlc, protoc-gen-go). These will be installed to `$GOPATH\bin`. You
can install them to the `bin` directory by running
`GOBIN=$(pwd)/bin make install-tools`.

## Tests

Run `make test` to run the tests

## Linting

Run `make lint` to run `golangci-lint`. Run `make lint-changes` to run
`golangci-lint` and show newly introduces linter warnings.

## Code generation

Run `make protoc` to generate protocol buffers related files. `make generate`
will generate sqlc related files.

## Running

Run `make run` to start shuttermint.
