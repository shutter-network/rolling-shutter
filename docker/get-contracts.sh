#!/usr/bin/env bash
## Needs to be bash, for the variable expansion to work!
source ./common.sh
set -e

CONTRACTS_JSON=$(jq '.transactions[]|(select(.function==null))|{(.contractName|tostring): .contractAddress}' data/deployments/Deploy.gnosh.s.sol/1337/run-latest.json)
#echo ${CONTRACTS_JSON} | jq -r ".[]|to_entries"


for s in $(echo ${CONTRACTS_JSON} | jq -r "to_entries|map(\"\(.key)=\(.value|tostring)\")| .[] " ); do
	export $s
done

for cfg in keyper-{0..3}.toml;
do
  config_path=config/${cfg};
  echo $config_path

  for name in KeyperSetManager KeyperSet KeyBroadcastContract Sequencer ValidatorRegistry;
  do
	key=$name
	value="${!name}"
	${BB} sed -i "/^$key =/c$key = \"$value\"" "${config_path}"
  done
done
