#!/usr/bin/env bash

source ./common.sh
set +ex

echo "Testing decryption key generation"
EPOCH_ID=$(LC_ALL=C tr -dc 'a-f0-9' </dev/urandom | head -c64)
json_body="{\"jsonrpc\": \"2.0\", \"method\": \"get_decryption_key\", \"id\": 1, \"params\": [\"1\", \"${EPOCH_ID}\"]}"
curl -XGET http://localhost:8754/api/v1/rpc -d "${json_body}"
sleep 5
DECRYPTION_KEY_MSGS=$(${DC} logs snapshot | grep ${EPOCH_ID} | grep -c decryptionKey)
if [[ $DECRYPTION_KEY_MSGS -gt 0 ]]; then
    echo "Decryption successful with $DECRYPTION_KEY_MSGS / 3 nodes"
    exit 0
else
    echo "Decryption failed"
    exit 1
fi
