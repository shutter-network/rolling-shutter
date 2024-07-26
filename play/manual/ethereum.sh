#!/bin/bash

set -e

private_key=
rpc_url=https://rpc.chiadochain.net/

keypers=0x9A1ba2D523AAB8f7784870B639924103d25Bb714,0x7b79Ba0f76eE49F6246c0034A2a3445C281a67EB,0x62F6DC5638250bD9edE84DFBfa54efA263186a4a
threshold=2
activation_delta=60

contracts_root_dir=../../../
shop_contracts_dir=shop-contracts
gnosh_contracts_dir=gnosh-contracts

cd $contracts_root_dir

# deploy keyper set manager and key broadcast contract
cd $shop_contracts_dir
function deploy_shop() {
    export PRIVATE_KEY=$private_key
    forge script script/DeployGnosis.s.sol --rpc-url $rpc_url --broadcast
}
deploy_shop_output=$(deploy_shop)

# extract contract addresses
extract_address() {
    local contract_name="$1"
    local output="$2"
    local address=""
    address=$(echo "$output" | grep -o "$contract_name: 0x[[:xdigit:]]\{40\}" | cut -d' ' -f2)
    echo "$address"
}

keyper_set_manager_address=$(extract_address "keyperSetManager" "$deploy_shop_output")
key_broadcast_address=$(extract_address "keyBroadcastContract" "$deploy_shop_output")

function add_empty_keyper_set() {
    export PRIVATE_KEY=$private_key
    export KEYPER_ADDRESSES=
    export THRESHOLD=0
    export ACTIVATION_DELTA=$activation_delta
    export KEYPERSETMANAGER_ADDRESS=$keyper_set_manager_address
    export KEYBROADCAST_ADDRESS=$key_broadcast_address
    forge script script/AddKeyperSet.s.sol --rpc-url $rpc_url --broadcast
}
add_empty_keyper_set

# add keyper set
function add_keyper_set() {
    export PRIVATE_KEY=$private_key
    export KEYPER_ADDRESSES=$keypers
    export THRESHOLD=$threshold
    export ACTIVATION_DELTA=$activation_delta
    export KEYPERSETMANAGER_ADDRESS=$keyper_set_manager_address
    export KEYBROADCAST_ADDRESS=$key_broadcast_address
    forge script script/AddKeyperSet.s.sol --rpc-url $rpc_url --broadcast
}
add_keyper_set_output=$(add_keyper_set)

eon_key_publish_address=$(extract_address "eonKeyPublish" "$add_keyper_set_output")

# deploy sequencer contract
cd ../$gnosh_contracts_dir
function deploy_gnosh() {
    export DEPLOY_KEY=$PRIVATE_KEY
    export ETHERSCAN_API_KEY=
    forge script script/deploySequencer.s.sol --rpc-url $rpc_url --broadcast
}
deploy_gnosh_output=$(deploy_gnosh)

sequencer_address=$(extract_address "Sequencer" "$deploy_gnosh_output")
validator_registry_address=$(extract_address "ValidatorRegistry" "$deploy_gnosh_output")

echo "KeyperSetManager = '$keyper_set_manager_address'"
echo "KeyBroadcastContract = '$key_broadcast_address'"
echo "Sequencer = '$sequencer_address'"
echo "ValidatorRegistry = '$validator_registry_address'"
echo "EonKeyPublish = '$eon_key_publish_address'"

echo "contracts deployed and configured"
