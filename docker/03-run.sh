#!/usr/bin/env bash
set -xe

if docker compose ls >/dev/null 2>&1; then
  DC="docker compose"
else
  DC=docker-compose
fi

$DC stop node
$DC rm -f node
rm -rf data/deployments

$DC up -d node
sleep 20
$DC up -d chain
echo "We need to wait for the chain to reach height >= 1"
sleep 25
echo "This will take a while..."
$DC run --rm --no-deps --entrypoint /rolling-shutter chain bootstrap \
  --deployment-dir /deployments/localhost \
  --ethereum-url http://node:8545 \
  --shuttermint-url http://chain:26657 \
  --signing-key 479968ffa5ee4c84514a477a8f15f3db0413964fd4c20b08a55fed9fed790fad

echo "Starting entire system"

$DC up -d
sleep 5
$DC status
