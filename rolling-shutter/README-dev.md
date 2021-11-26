# shuttermint

## Installation

Make sure you have at least go version 1.16 installed. Make sure `PATH` contains
`$GOPATH/bin`. If `GOPATH` isn't set, it defaults to `${HOME}/go`. Install
protoc version v3.18.0 from
https://github.com/protocolbuffers/protobuf/releases/

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
`golangci-lint` and show newly introduced linter warnings.

## Code generation

Run `make generate` to generate sqlc and protocol buffers related files.

## Running

Run `make run` to start shuttermint.

# Managing tools with asdf

[asdf](https://github.com/asdf-vm/asdf) can be used to install and manage the
versions of the different tools we use to build rolling-shutter. Please follow
the [installation guide](https://asdf-vm.com/guide/getting-started.html).

When asdf is ready, install the following plugins. In case you don't want to
manage some of the tools with asdf, skip installation of the corresponding
plugins.

```
asdf plugin add babashka
asdf plugin add circleci https://github.com/trnubo/asdf-circleci.git
asdf plugin add clojure
asdf plugin add golang
asdf plugin add java
asdf plugin add nodejs
asdf plugin add protoc
```

Finally, install the tools by running the following inside the rolling shutter
git repository:

```
asdf install
```
