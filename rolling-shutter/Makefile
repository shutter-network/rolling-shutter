.POSIX:

GO		?= go
GOFLAGS		?=
BINDIR		?= ./bin
EXECUTABLE	?= ${BINDIR}/rolling-shutter

build:
	@VERSION=`git describe --tags --always --abbrev=4 --dirty --match 'rolling-shutter/v*' | sed -e s#^rolling-shutter/##`; \
	echo "Building ${EXECUTABLE} $${VERSION}"; \
	${GO} build ${GOFLAGS} -o ${EXECUTABLE} -ldflags "-X github.com/shutter-network/shutter/shuttermint/cmd/shversion.version=$${VERSION}"; \

wasm:
	GOARCH=wasm GOOS=js ${GO} build ${GOFLAGS} -o ${BINDIR}/shcrypto.wasm ./shcryptowasm/main_wasm.go

protoc:
	${GO} generate ./shmsg/

test-unit:
	@echo "====================> Running unit tests"
	gotestsum -- -short ${GOFLAGS} ./...

test-integration:
	@echo "====================>  Running integration tests"
	gotestsum -- -p 1 -run Integration -count=1 ${GOFLAGS} ./...

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
	${GO} install github.com/golangci/golangci-lint/cmd/golangci-lint

install-cobra:
	${GO} install github.com/spf13/cobra/cobra@latest

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

.PHONY: build clean test test-all test-unit test-integration generate install-abigen install-protoc-gen-go install-oapi-codegen install-golangci-lint install-cobra install-gofumpt install-gotestsum install-tools lint lint-changes coverage abigen
