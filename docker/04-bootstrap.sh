#!/usr/bin/env bash

source ./common.sh

echo "Submitting bootstrap transaction"

$DC run --rm --no-deps --entrypoint /rolling-shutter chain-0-validator bootstrap \
    --deployment-dir /deployments/dockerGeth \
    --ethereum-url http://geth:8545 \
    --shuttermint-url http://chain-0-sentry:${TM_RPC_PORT} \
    --signing-key 479968ffa5ee4c84514a477a8f15f3db0413964fd4c20b08a55fed9fed790fad
