#!/usr/bin/env bash

source ./common.sh

echo "Submitting bootstrap transaction"

$DC run --rm --no-deps --entrypoint /rolling-shutter chain-0-validator op-bootstrap fetch-keyperset \
    --config /config/op-bootstrap.toml
    # --deployment-dir /deployments/dockerGeth \
