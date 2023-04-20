#!/usr/bin/env bash
set -xe

if docker compose ls >/dev/null 2>&1; then
  DC="docker compose"
else
  DC=docker-compose
fi

$DC stop db
$DC rm -f db

rm -rf data/db

$DC up -d db
sleep 40

for cmd in collator snapshot keyper-0 keyper-1 keyper-2; do
  $DC exec db createdb -U postgres $cmd
  $DC run --rm --no-deps $cmd initdb --config /config/${cmd}.toml
done

$DC stop db
