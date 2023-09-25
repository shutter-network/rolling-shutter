#!/usr/bin/env bash

source ./common.sh

$DC stop geth
$DC rm -f geth
$DC stop chain-{0..3}-validator chain-seed
$DC rm -f chain-{0..3}-validator chain-seed

${BB} rm -rf data/geth
${BB} rm -rf data/chain-{0..3}-validator data/chain-seed
${BB} mkdir -p data/chain-{0..3}-validator/config data/chain-seed/config
${BB} chmod -R a+rwX data/chain-{0..3}-validator/config data/chain-seed/config
${BB} rm -rf data/deployments

# has geth as dependency
$DC up deploy-contracts

$DC run --rm --no-deps chain-seed init \
    --root /chain \
    --blocktime 1 \
    --listen-address tcp://0.0.0.0:${TM_RPC_PORT} \
    --role seed

for num in 0 1 2 3; do
    validator_cmd=chain-$num-validator

    $DC run --rm --no-deps ${validator_cmd} init \
        --root /chain \
        --genesis-keyper 0x440Dc6F164e9241F04d282215ceF2780cd0B755e \
        --blocktime 1 \
        --listen-address tcp://0.0.0.0:${TM_RPC_PORT} \
        --role validator

    ${BB} sed -i "/ValidatorPublicKey/c\ValidatorPublicKey = \"$(cat data/${validator_cmd}/config/priv_validator_pubkey.hex)\"" /config/keyper-${num}.toml

    if [ $num -eq 0 ]; then
        for destination in data/chain-seed/config/ data/chain-{1..3}-validator/config/ ; do
            ${BB} cp -v data/chain-0-validator/config/genesis.json "${destination}"
        done
    fi
done

seed_node=$(cat data/chain-seed/config/node_key.json.id)@chain-seed:${TM_P2P_PORT}
validator_0_node=$(cat data/chain-0-validator/config/node_key.json.id)@chain-0-validator:${TM_P2P_PORT}

${BB} sed -i "/^moniker/c\moniker = \"chain-seed\"" data/chain-seed/config/config.toml

for num in {0..3}; do
    validator_cmd=chain-$num-validator

    validator_id=$(cat data/${validator_cmd}/config/node_key.json.id)
    validator_node=${validator_id}@${validator_cmd}:${TM_P2P_PORT}

    # set seed node for validator
    ${BB} sed -i "/^bootstrap_peers =/c\bootstrap_peers = \"${validator_0_node}\"" data/${validator_cmd}/config/config.toml
    ${BB} sed -i "/^external_address =/c\external_address = \"${validator_cmd}:${TM_P2P_PORT}\"" data/${validator_cmd}/config/config.toml
    ${BB} sed -i "/^pex/c\pex = true" data/${validator_cmd}/config/config.toml
    ${BB} sed -i "/^moniker/c\moniker = \"${validator_cmd}\"" data/${validator_cmd}/config/config.toml

done

$DC stop -t 30
