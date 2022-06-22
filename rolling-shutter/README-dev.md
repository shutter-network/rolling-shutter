# shuttermint

## Installation

Make sure you have at least go version 1.18 installed. Make sure `PATH` contains
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

# Managing tools with asdf

[asdf](https://github.com/asdf-vm/asdf) can be used to install and manage the
versions of the different tools we use to build rolling-shutter. Please follow
the [installation guide](https://asdf-vm.com/guide/getting-started.html) for
detailed instructions. On an debian/ubuntu based system the following should
work for a user using bash:

```
sudo apt install curl git build-essential unzip
git clone https://github.com/asdf-vm/asdf.git ~/.asdf --branch v0.9.0
```

And tell bash to source asdf's init file by appending these two lines to
~/.bashrc:

```
. $HOME/.asdf/asdf.sh
. $HOME/.asdf/completions/asdf.bash
```

When asdf is ready, install the needed plugins. In case you don't want to manage
some of the tools with asdf, skip installation of the corresponding plugins.

```
asdf plugin add babashka
asdf plugin add circleci https://github.com/trnubo/asdf-circleci.git
asdf plugin add golang
asdf plugin add golangci-lint
asdf plugin add nodejs
asdf plugin add protoc
asdf plugin add sqlc https://github.com/schmir/asdf-sqlc.git
asdf plugin add solidity
asdf plugin add tinygo https://github.com/schmir/asdf-tinygo.git
asdf plugin add binaryen https://github.com/birros/asdf-binaryen.git
asdf plugin add pre-commit
```

Finally, install the tools by running the following inside the rolling shutter
git repository:

```
asdf install
```

If you want to run the whole system tests inside the play directory, you also
have to install clojure and java:

```
asdf plugin add java
asdf plugin add clojure
asdf install java
asdf install clojure
```
