.POSIX:

GO		?= go
GOFLAGS		?=
TINYGO		?= tinygo
WASMOPT     ?= wasm-opt
BINDIR		?= ./bin
EXECUTABLE	?= ${BINDIR}/rolling-shutter

build:
	${GO} build ${GOFLAGS} -o ${EXECUTABLE}

shcryptowasm: wasm
	echo "The 'shcryptowasm' target is deprecated, use 'wasm' instead."

wasm:
	${TINYGO} build -target wasm -size full -o ${BINDIR}/shutter-crypto.unopt.wasm ./shcryptowasm/shutter_crypto_wasm.go
	${WASMOPT} -Oz --vacuum --strip-debug --strip-dwarf --strip-producers --strip-target-features -o ${BINDIR}/shutter-crypto.wasm ${BINDIR}/shutter-crypto.unopt.wasm
	rm ${BINDIR}/shutter-crypto.unopt.wasm

wasm-js: wasm
	$(MAKE) -C ../js/shutter-crypto build

protoc:
	${GO} generate ./shmsg/

test-unit:
	@echo "====================> Running unit tests"
	gotestsum -- -race -short ${GOFLAGS} ./...

test-integration:
	@echo "====================>  Running integration tests"
	gotestsum -- -race -p 1 -run Integration -count=1 ${GOFLAGS} ./...

test: test-unit

test-all: test-unit test-integration

generate:
	${GO} generate -x ./...

coverage:
	${GO} test ${GOFLAGS} -covermode=count -coverprofile=coverage.out ./...
	${GO} tool cover -html=coverage.out

clean:
	rm -f ${EXECUTABLE}

install-tools: install-abigen install-sqlc install-protoc-gen-go  install-oapi-codegen install-golangci-lint install-cobra install-gofumpt install-stringer install-gci install-gotestsum

install-sqlc:
	${GO} install github.com/kyleconroy/sqlc/cmd/sqlc

install-abigen:
	${GO} install github.com/ethereum/go-ethereum/cmd/abigen

install-protoc-gen-go:
	${GO} install google.golang.org/protobuf/cmd/protoc-gen-go

install-oapi-codegen:
	${GO} install github.com/deepmap/oapi-codegen/cmd/oapi-codegen

install-golangci-lint:
	${GO} install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

install-cobra:
	${GO} install github.com/spf13/cobra-cli@latest

install-gofumpt:
	${GO} install mvdan.cc/gofumpt@latest

install-stringer:
	${GO} install golang.org/x/tools/cmd/stringer

install-gci:
	${GO} install github.com/daixiang0/gci@latest

install-gotestsum:
	${GO} install gotest.tools/gotestsum@latest

lint:
	golangci-lint run --tests

lint-changes:
	base=`git merge-base HEAD origin/main`; \
	golangci-lint run --new-from-rev $${base}

abigen:
	go generate -x ./contract

.PHONY: build clean test test-all test-unit test-integration generate install-abigen install-protoc-gen-go install-oapi-codegen install-golangci-lint install-cobra install-gofumpt install-gotestsum install-tools lint lint-changes coverage abigen shcryptowasm wasm wasm-js wasm-legacy
