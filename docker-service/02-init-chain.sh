#!/usr/bin/env bash

source ./common.sh

$DC stop blockchain
$DC rm -f blockchain
$DC stop chain-{0..2}-validator chain-seed
$DC rm -f chain-{0..2}-validator chain-seed

${BB} rm -rf data/chain-{0..2}-validator data/chain-seed
${BB} mkdir -p data/chain-{0..2}-validator/config data/chain-seed/config
${BB} chmod -R a+rwX data/chain-{0..2}-validator/config data/chain-seed/config
${BB} rm -rf data/deployments

# has blockchain as dependency
$DC up -d blockchain
sleep 5
$DC up deploy-contracts

bash get-contracts.sh
# setup chain-seed
$DC run --rm --no-deps chain-seed init \
    --root /chain \
    --blocktime 1 \
    --listen-address tcp://0.0.0.0:${TM_RPC_PORT} \
    --role seed

seed_node=$(cat data/chain-seed/config/node_key.json.id)@chain-seed:${TM_P2P_PORT}

${BB} sed -i "/^moniker/c\moniker = \"chain-seed\"" data/chain-seed/config/config.toml

# configure validators and keypers 0-2
for num in {0..2}; do
    validator_cmd=chain-$num-validator

    $DC run --rm --no-deps ${validator_cmd} init \
        --root /chain \
        --genesis-keyper 0x346a9357D8EB6F0FbC4894ed6DBb1eCCA1051c09 \
        --blocktime 1 \
        --listen-address tcp://0.0.0.0:${TM_RPC_PORT} \
        --role validator

    validator_id=$(cat data/${validator_cmd}/config/node_key.json.id)
    validator_node=${validator_id}@${validator_cmd}:${TM_P2P_PORT}
    validator_config_path=data/${validator_cmd}/config/config.toml

    # share genesis
    if [ $num -eq 0 ]; then
        for destination in data/chain-seed/config/ data/chain-{1..2}-validator/config/; do
            ${BB} cp -v data/chain-0-validator/config/genesis.json "${destination}"
        done
    fi

    # set validator publickey for keyper
    ${BB} sed -i "/ValidatorPublicKey/c\ValidatorPublicKey = \"$(cat data/${validator_cmd}/config/priv_validator_pubkey.hex)\"" /config/keyper-${num}.toml

    # set seed node for chain bootstrap
    ${BB} sed -i "/^seeds =/c\seeds = \"${seed_node}\"" "${validator_config_path}"
    # fix external address for docker internal communication
    ${BB} sed -i "/^external_address =/c\external_address = \"${validator_cmd}:${TM_P2P_PORT}\"" "${validator_config_path}"
    # give a nice name
    ${BB} sed -i "/^moniker/c\moniker = \"${validator_cmd}\"" "${validator_config_path}"

done

$DC stop -t 30
