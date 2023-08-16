#!/usr/bin/env bash

source ./common.sh

$DC stop db
$DC rm -f db

${BB} rm -rf data/db

$DC up -d db
$DC run --rm --no-deps dockerize -wait tcp://db:5432 -timeout 40s

for cmd in snapshot keyper-0 keyper-1 keyper-2 keyper-3; do
    $DC exec db createdb -U postgres $cmd
    $DC run --rm --no-deps $cmd initdb --config /config/${cmd}.toml
done

$DC stop db
