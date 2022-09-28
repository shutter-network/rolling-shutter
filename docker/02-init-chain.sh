#!/usr/bin/env bash
set -xe

if docker compose ls >/dev/null 2>&1; then
  # compose v2
  DC="docker compose"
else
  DC=docker-compose
fi

$DC stop geth
$DC rm -f geth
$DC stop chain
$DC rm -f chain

rm -rf data/geth
rm -rf data/chain
rm -rf data/deployments

$DC up deploy-contracts  # has geth as dependency

$DC run --rm --no-deps chain init \
  --root /chain \
  --genesis-keyper 0x440Dc6F164e9241F04d282215ceF2780cd0B755e \
  --dev \
  --blocktime 5 \
  --listen-address tcp://0.0.0.0:26657

$DC up -d chain
echo "We need to wait for the chain to reach height >= 1"
sleep 25
echo "This will take a while..."
$DC run --rm --no-deps --entrypoint /rolling-shutter chain bootstrap \
  --deployment-dir /deployments/dockerGeth \
  --ethereum-url http://geth:8545 \
  --shuttermint-url http://chain:26657 \
  --signing-key 479968ffa5ee4c84514a477a8f15f3db0413964fd4c20b08a55fed9fed790fad

$DC stop geth chain
