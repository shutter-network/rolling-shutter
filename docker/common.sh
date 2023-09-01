BB="docker run --rm -v $(pwd)/data:/data -v $(pwd)/config:/config -w / busybox"

if docker compose ls >/dev/null 2>&1; then
    DC="docker compose"
else
    DC=docker-compose
fi

set -xe
