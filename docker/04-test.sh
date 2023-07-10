#!/usr/bin/env bash
set -xe

if docker compose ls >/dev/null 2>&1; then
  DC="docker compose"
else
  DC=docker-compose
fi

echo "Testing decryption key generation"
EPOCH_ID="480184f2b2dedec2641fb1a0b8cb1f0a8af8e7edd90f2f5acfc0858c29ed964c"
json_body="{\"jsonrpc\": \"2.0\", \"method\": \"get_decryption_key\", \"id\": 1, \"params\": [\"1\", \"${EPOCH_ID}\"]}"
curl -XGET http://localhost:8754/api/v1/rpc -d "${json_body}"
sleep 3
DECRYPTION_KEY_MSGS=$(${DC} logs snapshot|grep ${EPOCH_ID}|grep -c decryptionKey)
if [ $DECRYPTION_KEY_MSGS -eq 3 ]; then
  echo "Decryption successful"
  exit 0
else
  echo "Decryption failed"
  exit 1
fi
