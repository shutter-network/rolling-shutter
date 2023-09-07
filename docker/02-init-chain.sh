#!/usr/bin/env bash

source ./common.sh

$DC stop geth
$DC rm -f geth
$DC stop chain-{0..3}-{validator,sentry} chain-seed
$DC rm -f chain-{0..3}-{validator,sentry} chain-seed

${BB} rm -rf data/geth
${BB} rm -rf data/chain-{0..3}-{validator,sentry} data/chain-seed
${BB} mkdir -p data/chain-{0..3}-{validator,sentry}/config data/chain-seed/config
${BB} chmod -R a+rwX data/chain-{0..3}-{validator,sentry}/config data/chain-seed/config
${BB} rm -rf data/deployments

# has geth as dependency
$DC up deploy-contracts

TM_P2P_PORT=26656
TM_RPC_PORT=26657

$DC run --rm --no-deps chain-seed init \
    --root /chain \
    --blocktime 1 \
    --listen-address tcp://127.0.0.1:${TM_RPC_PORT} \
    --role seed

for num in 0 1 2 3; do
    validator_cmd=chain-$num-validator
    sentry_cmd=chain-$num-sentry

    $DC run --rm --no-deps ${sentry_cmd} init \
        --root /chain \
        --blocktime 1 \
        --listen-address tcp://0.0.0.0:${TM_RPC_PORT} \
        --role sentry

    # TODO: check if validator can have listen-address tcp://127.0.0.1...
    $DC run --rm --no-deps ${validator_cmd} init \
        --root /chain \
        --genesis-keyper 0x440Dc6F164e9241F04d282215ceF2780cd0B755e \
        --blocktime 1 \
        --listen-address tcp://127.0.0.1:${TM_RPC_PORT} \
        --role validator

    ${BB} sed -i "/ValidatorPublicKey/c\ValidatorPublicKey = \"$(cat data/${validator_cmd}/config/priv_validator_pubkey.hex)\"" /config/keyper-${num}.toml

    if [ $num -eq 0 ]; then
        for destination in data/chain-seed/config/ data/chain-{1..3}-validator/config/ data/chain-{0..3}-sentry/config/; do
            ${BB} cp -v data/chain-0-validator/config/genesis.json "${destination}"
        done
    fi
done

seed_node=$(cat data/chain-seed/config/node_key.json.id)@chain-seed:${TM_P2P_PORT}

for num in 0 1 2 3; do
    sentry_cmd=chain-$num-sentry
    validator_cmd=chain-$num-validator

    validator_id=$(cat data/${validator_cmd}/config/node_key.json.id)
    validator_node=${validator_id}@${validator_cmd}:${TM_P2P_PORT}
    sentry_node=$(cat data/${sentry_cmd}/config/node_key.json.id)@${sentry_cmd}:${TM_P2P_PORT}

    # set seed node for sentry
    ${BB} sed -i "/^persistent-peers =/c\persistent-peers = \"${seed_node}\"" data/${sentry_cmd}/config/config.toml
    # set validator node for sentry
    ${BB} sed -i "/^private-peer-ids =/c\private-peer-ids = \"${validator_id}\"" data/${sentry_cmd}/config/config.toml
    ${BB} sed -i "/^unconditional-peer-ids =/c\unconditional-peer-ids = \"${validator_id}\"" data/${sentry_cmd}/config/config.toml
    ${BB} sed -i "/^external-address =/c\external-address = \"${sentry_cmd}:${TM_P2P_PORT}\"" data/${sentry_cmd}/config/config.toml

    # set sentry node for validator
    ${BB} sed -i "/^persistent-peers =/c\persistent-peers = \"${sentry_node}\"" data/${validator_cmd}/config/config.toml
    ${BB} sed -i "/^external-address =/c\external-address = \"${validator_cmd}:${TM_P2P_PORT}\"" data/${validator_cmd}/config/config.toml
done

$DC up -d chain-seed chain-{0..3}-{sentry,validator} keyper-{0..3}

echo "We need to wait for the chain to reach height >= 1"
sleep 5

$DC run --rm --no-deps --entrypoint /rolling-shutter chain-0-validator bootstrap \
    --deployment-dir /deployments/dockerGeth \
    --ethereum-url http://geth:8545 \
    --shuttermint-url http://chain-0-sentry:${TM_RPC_PORT} \
    --signing-key 479968ffa5ee4c84514a477a8f15f3db0413964fd4c20b08a55fed9fed790fad

$DC stop -t 30 geth chain-seed chain-{0..3}-{sentry,validator} keyper-{0..3}
