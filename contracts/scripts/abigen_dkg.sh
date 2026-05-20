#!/bin/bash
#
# Generate Go bindings for the DKGContract and ECIESKeyRegistry contracts.
#
# The contracts live in the top-level foundry-based contracts repo
# (../../../contracts relative to this script). We build them with forge and
# then run abigen against the resulting artifacts, producing a single
# binding_dkg.abigen.gen.go file in the contract package.

set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONTRACTS_DIR="$(cd "$SCRIPT_DIR/../../../contracts" && pwd)"
OUT_GO="$(cd "$SCRIPT_DIR/../../rolling-shutter/contract" && pwd)/binding_dkg.abigen.gen.go"

(cd "$CONTRACTS_DIR" && forge build --offline)

TMP_COMBINED="$(mktemp --suffix=.json)"
trap 'rm -f "$TMP_COMBINED"' EXIT

jq -n \
    --slurpfile dkg "$CONTRACTS_DIR/out/DKGContract.sol/DKGContract.json" \
    --slurpfile ecies "$CONTRACTS_DIR/out/ECIESKeyRegistry.sol/ECIESKeyRegistry.json" \
'{
  contracts: {
    "src/common/DKGContract.sol:DKGContract": {
      abi: ($dkg[0].abi | tostring),
      bin: $dkg[0].bytecode.object,
      "bin-runtime": $dkg[0].deployedBytecode.object,
      userdoc: "{}",
      devdoc: "{}"
    },
    "src/common/ECIESKeyRegistry.sol:ECIESKeyRegistry": {
      abi: ($ecies[0].abi | tostring),
      bin: $ecies[0].bytecode.object,
      "bin-runtime": $ecies[0].deployedBytecode.object,
      userdoc: "{}",
      devdoc: "{}"
    }
  },
  version: "foundry"
}' > "$TMP_COMBINED"

abigen \
    --combined-json "$TMP_COMBINED" \
    --pkg contract \
    --out "$OUT_GO"
