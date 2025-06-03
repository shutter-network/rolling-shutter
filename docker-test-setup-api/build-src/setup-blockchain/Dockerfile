# syntax=docker/dockerfile:1.4
FROM --platform=linux/amd64 ghcr.io/foundry-rs/foundry

WORKDIR /anvil

ENTRYPOINT anvil --host 0.0.0.0 -p 8545 -b 2 --dump-state /data/blockchain.json