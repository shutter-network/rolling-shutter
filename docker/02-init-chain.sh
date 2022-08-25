#!/usr/bin/env bash
set -xe

if docker compose ls >/dev/null 2>&1; then
  # compose v2
  DC="docker compose"
else
  DC=docker-compose
fi

$DC stop chain
$DC rm -f chain

rm -rf data/chain

$DC run --rm --no-deps chain init \
  --root /chain \
  --genesis-keyper 0x440Dc6F164e9241F04d282215ceF2780cd0B755e \
  --dev \
  --blocktime 5 \
  --listen-address tcp://0.0.0.0:26657
