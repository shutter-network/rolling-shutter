#!/bin/bash

export ROLLING_SHUTTER_BOOTSTRAP_SIGNING_KEY=479968ffa5ee4c84514a477a8f15f3db0413964fd4c20b08a55fed9fed790fad
export ROLLING_SHUTTER_CHAIN_GENESIS_KEYPER=0x440Dc6F164e9241F04d282215ceF2780cd0B755e

(cd .. && bb init:testchain)
(cd .. && bb chain) &

sleep 2
rm -f keyperset.json
./rolling-shutter op-bootstrap fetch-keyperset --config op-bootstrap-config.toml
./rolling-shutter op-bootstrap --config op-bootstrap-config.toml

wait
