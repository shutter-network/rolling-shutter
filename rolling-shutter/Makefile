.POSIX:

GO		?= go
GOFLAGS		?=
TINYGO		?= tinygo
WASMOPT     ?= wasm-opt
BINDIR		?= ./bin
WASMDIR		?= ../js/shutter-crypto
EXECUTABLE	?= ${BINDIR}/rolling-shutter
GOPATH          ?= $(${GO} env GOPATH)

build:
	${GO} build ${GOFLAGS} -ldflags="-X github.com/shutter-network/rolling-shutter/rolling-shutter/cmd/shversion.version=$(shell git describe --always --dirty)" -o ${EXECUTABLE}

shcryptowasm: wasm
	echo "The 'shcryptowasm' target is deprecated, use 'wasm' instead."

wasm:
	${TINYGO} build -target wasm -size full -o ${BINDIR}/shutter-crypto.unopt.wasm ./shcryptowasm/shutter_crypto_wasm.go
	${WASMOPT} -Oz --vacuum --strip-debug --strip-dwarf --strip-producers --strip-target-features -o ${BINDIR}/shutter-crypto.wasm ${BINDIR}/shutter-crypto.unopt.wasm
	rm ${BINDIR}/shutter-crypto.unopt.wasm
	rm -rf ${WASMDIR}/derived
	mkdir -p ${WASMDIR}/derived
	rsync ${BINDIR}/shutter-crypto.wasm ${WASMDIR}/derived/shutter-crypto.wasm

wasm-js: wasm
	$(MAKE) -C ${WASMDIR} build

wasm-js-test: wasm
	$(MAKE) -C ${WASMDIR} test

protoc:
	${GO} generate ./shmsg/ ./p2pmsg

test-unit:
	@echo "====================> Running unit tests"
	gotestsum -- -race -short ${GOFLAGS} ./...

test-integration:
	@echo "====================>  Running integration tests"
	gotestsum -- -race -p 1 -run Integration -count=1 ${GOFLAGS} ./...

test: test-unit

test-all: test-unit test-integration

generate: install-codegen-tools
	${GO} generate -skip="make docs" -x ./...
	${GO} generate -run="make docs" -x ./...

coverage:
	${GO} test ${GOFLAGS} -covermode=count -coverprofile=coverage.out ./...
	${GO} tool cover -html=coverage.out

clean:
	rm -f ${EXECUTABLE}

install-tools: install-codegen-tools install-golangci-lint install-cobra install-gofumpt install-gci install-gotestsum

# code generation tools: pin version
install-codegen-tools: install-npm install-abigen install-sqlc install-protoc-gen-go  install-oapi-codegen install-go-enum

../contracts/node_modules/.mark-npm-install: ../contracts/package.json ../contracts/package-lock.json
	cd ../contracts && npm install
	@touch ../contracts/node_modules/.mark-npm-install

install-npm: ../contracts/node_modules/.mark-npm-install

${GOPATH}/bin/sqlc:
	${GO} install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.22.0
install-sqlc: ${GOPATH}/bin/sqlc


${GOPATH}/bin/go-enum:
	${GO} install github.com/abice/go-enum@v0.5.6
install-go-enum: ${GOPATH}/bin/go-enum

${GOPATH}/bin/abigen:
	${GO} install github.com/ethereum/go-ethereum/cmd/abigen@v1.12.0
install-abigen: ${GOPATH}/bin/abigen


${GOPATH}/bin/protoc-gen-go:
	${GO} install google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0
install-protoc-gen-go: ${GOPATH}/bin/protoc-gen-go


${GOPATH}/bin/oapi-codegen:
	${GO} install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.1
install-oapi-codegen: ${GOPATH}/bin/oapi-codegen

# non code generation tools
install-golangci-lint:
	${GO} install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2

install-gofumpt:
	${GO} install mvdan.cc/gofumpt@latest

install-gci:
	${GO} install github.com/daixiang0/gci@latest

install-gotestsum:
	${GO} install gotest.tools/gotestsum@latest

install-asdf-plugins:
	../tools/asdf-install-plugins.sh

install-asdf: install-asdf-plugins
	@# Various packages installed by asdf internally also use make and don't like if an external make has already
	@# tinkered with then environment.
	@# Therefore we unset MAKELEVEL here.
	@env -u MAKELEVEL asdf install

lint:
	golangci-lint run --tests

lint-changes:
	base=`git merge-base HEAD origin/main`; \
	golangci-lint run --new-from-rev $${base}

abigen:
	go generate -x ./contract

.PHONY: build clean test test-all test-unit test-integration generate install-codegen-tools install-abigen install-protoc-gen-go install-oapi-codegen install-golangci-lint install-cobra install-gofumpt install-gotestsum install-tools lint lint-changes coverage abigen shcryptowasm wasm wasm-js wasm-legacy
