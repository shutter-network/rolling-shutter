#!/usr/bin/env bash
## Needs to be bash, for the variable expansion to work!
source ./common.sh
source .env
set -e

CONTRACTS_JSON=$(jq '.transactions[]|(select(.function==null))|{(.contractName|tostring): .contractAddress}' data/deployments/Deploy.service.s.sol/31337/run-latest.json)
#echo ${CONTRACTS_JSON} | jq -r ".[]|to_entries"

for s in $(echo ${CONTRACTS_JSON} | jq -r "to_entries|map(\"\(.key)=\(.value|tostring)\")| .[] "); do
    export $s
done

for cfg in keyper-{0..2}.toml; do
    config_path=config/${cfg}
    echo $config_path

    for name in KeyperSetManager KeyperSet KeyBroadcastContract ShutterRegistry; do
        key=$name
        value="${!name}"
        ${BB} sed -i "/^$key =/c$key = \"$value\"" "${config_path}"
    done
done

echo "Setting up bootstrap.toml"
${BB} sed -i "/^KeyperSetManager =/cKeyperSetManager = \"${KeyperSetManager}\"" "config/bootstrap.toml"
${BB} sed -i "/^SigningKey =/cSigningKey = \"$(echo $DEPLOY_KEY | cut -b3-)\"" "config/bootstrap.toml"
