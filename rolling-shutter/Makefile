.POSIX:

PREFIX ?= /usr/local
GO ?= go
GOFLAGS ?=
SHROOT ?= $(CURDIR)
TESTROOT=$(SHROOT)/testchain
BINDIR ?= ./bin

EXECUTABLE ?= ${BINDIR}/rolling-shutter
OPENZEPPELIN_CONTRACTS=openzeppelin
OPENZEPPELIN_CONTRACTS_DIR=../contracts/openzeppelin

build:
	@VERSION=`git describe --tags --always --abbrev=4 --dirty`; \
	echo "Building ${EXECUTABLE} $${VERSION}"; \
	${GO} build ${GOFLAGS} -o ${EXECUTABLE} -ldflags "-X github.com/shutter-network/shutter/shuttermint/cmd/shversion.version=$${VERSION}"; \
	${GO} build ${GOFLAGS} -o ${BINDIR} -ldflags "-X github.com/shutter-network/shutter/shuttermint/cmd/shversion.version=$${VERSION}" ./sandbox/testclient

wasm:
	GOARCH=wasm GOOS=js ${GO} build ${GOFLAGS} -o ${BINDIR}/shcrypto.wasm ./shcryptowasm/main_wasm.go

protoc:
	protoc shmsg/*.proto --go_out=shmsg/

${TESTROOT}:
	${EXECUTABLE} init --dev --root ${TESTROOT}

init: ${TESTROOT}

run: build init
	${EXECUTABLE} chain --config ${TESTROOT}/config/config.toml

test:
	${GO} test ${GOFLAGS} ./...

generate:
	${GO} generate ./...

coverage:
	${GO} test ${GOFLAGS} -covermode=count -coverprofile=coverage.out ./...
	${GO} tool cover -html=coverage.out

clean:
	rm -rf ${TESTROOT}
	rm -f ${EXECUTABLE} ${BINDIR}/testclient ${BINDIR}/shuttermint

install:
	install -o root -m 0555 ${EXECUTABLE} ${PREFIX}/bin/

install-tools: install-protoc-gen-go install-golangci-lint install-cobra install-gofumpt install-stringer install-gci

install-protoc-gen-go:
	${GO} install google.golang.org/protobuf/cmd/protoc-gen-go

install-golangci-lint:
	${GO} install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

install-cobra:
	${GO} install github.com/spf13/cobra/cobra@latest

install-gofumpt:
	${GO} install mvdan.cc/gofumpt@latest

install-stringer:
	${GO} install golang.org/x/tools/cmd/stringer

install-gci:
	${GO} install github.com/daixiang0/gci@latest

lint:
	golangci-lint run --tests

lint-changes:
	base=`git merge-base HEAD origin/main`; \
	golangci-lint run --new-from-rev $${base}

.PHONY: build install run init clean test generate install-protoc-gen-go install-golangci-lint install-cobra install-gofumpt install-tools lint lint-changes coverage
