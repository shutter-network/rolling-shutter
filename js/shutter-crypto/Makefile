.POSIX:

NPM       ?= npm
NPX       ?= npx
TINYGO    ?= tinygo

build: derive
	${NPX} webpack build

publish: build
	${NPM} publish --access=public --tag=beta

derive: derived/shutter-crypto.wasm derived/wasm_exec.js

derived/shutter-crypto.wasm: ../../rolling-shutter/bin/shutter-crypto.wasm
	ln -sf ../../../rolling-shutter/bin/shutter-crypto.wasm derived/
derived/wasm_exec.js:
	cp $(shell ${TINYGO} env TINYGOROOT)/targets/wasm_exec.js derived/


.PHONY: build derive publish
