docs: rolling-shutter.md

rolling-shutter.md: ../medley/rootcmd/root.go $(shell find ../cmd -iname '*.go' ! -iname '*_test.go')
	go run gendocs.go
