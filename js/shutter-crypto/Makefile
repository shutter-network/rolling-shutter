.POSIX:

NPM       ?= npm
NPX       ?= npx
TINYGO    ?= tinygo

clean:
	rm -r dist
	mkdir -p dist

build: clean derive
	${NPX} webpack build
	rsync derived/shutter-crypto.wasm dist/shutter-crypto.wasm

publish: build
	${NPM} publish --access=public

pack: build
	${NPM} pack

derive: derived/shutter-crypto.wasm derived/wasm_exec.js

derived/shutter-crypto.wasm: ../../rolling-shutter/bin/shutter-crypto.wasm
	ln -sf ../../../rolling-shutter/bin/shutter-crypto.wasm derived/
derived/wasm_exec.js:
	cp $(shell ${TINYGO} env TINYGOROOT)/targets/wasm_exec.js derived/
	patch --quiet --directory derived --strip 1 --input ../patches/wasm_exec.patch

test:
	${NPM} test

.PHONY: build derive publish test
