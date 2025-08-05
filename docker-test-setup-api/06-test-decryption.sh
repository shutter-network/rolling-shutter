#!/usr/bin/env bash

source ./common.sh
set -e

echo "Submitting identity registration transaction"

CONTRACTS_JSON=$(jq '.transactions[]|(select(.function==null))|{(.contractName|tostring): .contractAddress}' data/deployments/Deploy.service.s.sol/31337/run-latest.json)

for s in $(echo ${CONTRACTS_JSON} | jq -r "to_entries|map(\"\(.key)=\(.value|tostring)\")| .[] "); do
    export $s
done

export TIMESTAMP=$(($(date +%s) + 50))
export IDENTITY_PREFIX=0x$(LC_ALL=C tr -dc 'a-f0-9' </dev/urandom | head -c64)
export REGISTRY_ADDRESS=${ShutterRegistry}
export EON=$($DC exec db psql -U postgres -d keyper-0 -t -c 'select max(eon) from eons;' | tr -d '[:space:]')
if [[ -z $EON ]]; then
    echo "No eonId found in keyper-0 db. Did you run bootstrap?"
    exit 1
fi

${DC} run --rm --no-deps register-identity
sleep 55

DECRYPTION_KEY_MSGS=$(${DC} logs keyper-0 | grep ${EON} | grep -c decryptionKey)
echo "DECRYPTION_KEY_MSGS: ${DECRYPTION_KEY_MSGS}"
if [[ $DECRYPTION_KEY_MSGS -gt 0 ]]; then
    echo "Decryption successful with $DECRYPTION_KEY_MSGS / 3 nodes"
    exit 0
else
    echo "Decryption failed"
    exit 1
fi
