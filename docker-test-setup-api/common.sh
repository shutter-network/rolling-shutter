BB="docker run --rm -v $(pwd)/data:/data -v $(pwd)/config:/config -w / busybox"
TM_P2P_PORT=26656
TM_RPC_PORT=26657

if docker compose ls >/dev/null 2>&1; then
    DC="docker compose"
else
    DC=docker-compose
fi

set -xe
