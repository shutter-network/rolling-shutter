#!/usr/bin/env sh

solc --combined-json abi,asm,ast,bin,bin-runtime,devdoc,function-debug,function-debug-runtime,generated-sources,generated-sources-runtime,hashes,metadata,opcodes,srcmap,srcmap-runtime,storage-layout,userdoc erc20.sol >erc20.combined.json

abigen --combined-json erc20.combined.json --pkg help --out erc20bindings.go

solc --combined-json abi,asm,ast,bin,bin-runtime,devdoc,function-debug,function-debug-runtime,generated-sources,generated-sources-runtime,hashes,metadata,opcodes,srcmap,srcmap-runtime,storage-layout,userdoc ShutterEventTrigger.sol >eventtrigger.combined.json

abigen --combined-json eventtrigger.combined.json --pkg help --out eventtriggerbindings.go
