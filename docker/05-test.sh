#!/usr/bin/env bash

source ./common.sh
set +ex

EPOCH_ID=$(LC_ALL=C tr -dc 'a-f0-9' </dev/urandom | head -c64)
EON_ID=$($DC exec db psql -U postgres -d keyper-0 -t -c 'select max(eon) from eons;' | tr -d '[:space:]')
if [[ -z $EON_ID ]]; then
    echo "No eonId found in keyper-0 db. Did you run bootstrap?"
    exit 1
fi
json_body="{\"jsonrpc\": \"2.0\", \"method\": \"get_decryption_key\", \"id\": 1, \"params\": [\"${EON_ID}\", \"${EPOCH_ID}\"]}"
echo "Testing decryption key generation for eonId ${EON_ID} and epoch ${EPOCH_ID}"
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
