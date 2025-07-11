#!/usr/bin/env sh

solc --combined-json abi,asm,ast,bin,bin-runtime,compact-format,devdoc,function-debug,function-debug-runtime,generated-sources,generated-sources-runtime,hashes,interface,metadata,opcodes,srcmap,srcmap-runtime,storage-layout,userdoc erc20.sol >erc20.combined.json

abigen --combined-json erc20.combined.json --pkg help --out ../erc20bindings.go
