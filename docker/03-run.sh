#!/usr/bin/env bash
set -xe

if docker compose ls >/dev/null 2>&1; then
  DC="docker compose"
else
  DC=docker-compose
fi

echo "Starting entire system"
$DC up -d keyper-0 keyper-1 keyper-2 collator snapshot caddy
sleep 5
$DC ps
