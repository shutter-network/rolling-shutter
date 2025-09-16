#!/usr/bin/env bash

source ./common.sh
source .env
set -e

CONTRACTS_JSON=$(jq '.transactions[]|(select(.function==null))|{(.contractName|tostring): .contractAddress}' data/deployments/Deploy.service.s.sol/31337/run-latest.json)

for s in $(echo ${CONTRACTS_JSON} | jq -r "to_entries|map(\"\(.key)=\(.value|tostring)\")| .[] "); do
    export $s
done

# Get keyper addresses from node-deploy.json
export KEYPER_ADDRESSES=$(jq -r '.keypers[0] | join(",")' config/node-deploy.json)

echo "Submitting Add Keyper Set transaction"
export THRESHOLD=2
export KEYPERSETMANAGER_ADDRESS=${KeyperSetManager}
export KEYBROADCAST_ADDRESS=${KeyBroadcastContract}
export ACTIVATION_DELTA=10

$DC run --rm --no-deps add-keyper-set
