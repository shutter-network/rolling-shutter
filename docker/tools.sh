#!/usr/bin/env bash
set -e

## Fund account

# DEPLOY PRIVATE KEY=44ea0c624dbec53682a11482f732dcd4e8581ed181fbfe2ad69e88523dc0a312
DEPLOY_ACCOUNT="0x346a9357D8EB6F0FbC4894ed6DBb1eCCA1051c09"

DEV_ACCOUNT=$(docker run --rm -it --network snapshutter_default curlimages/curl -Ss -H "Content-Type: application/json" -XPOST http://geth:8545 -d '{
        "jsonrpc":"2.0",
        "method":"eth_accounts",
        "params":[],
        "id":1
}'|jq -r ".result[0]")


DATA=$(jq -nc --arg dev "$DEV_ACCOUNT" --arg deploy "$DEPLOY_ACCOUNT" '{ jsonrpc:"2.0", method:"eth_sendTransaction", params:[{ from: $dev, to: $deploy, value: "0xc097ce7bc90715b34b9f1000000000" }], id:1}')

docker run --rm -it --network snapshutter_default curlimages/curl -H "Content-Type: application/json" -XPOST http://geth:8545 -d $DATA
