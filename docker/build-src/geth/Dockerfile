FROM ethereum/client-go:v1.14.12

RUN apk add --no-cache curl

HEALTHCHECK --interval=5s CMD curl -sSf -X POST http://127.0.0.1:8545 -H "Content-Type: application/json" --data-raw '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[], "id": 1}'
