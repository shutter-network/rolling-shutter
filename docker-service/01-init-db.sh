#!/usr/bin/env bash

source ./common.sh

mkdb() {
    $DC exec -T db createdb -U postgres $1
    $DC run -T --rm --no-deps $1 initdb --config /config/${1}.toml
}

$DC stop db
$DC rm -f db

${BB} rm -rf data/db

$DC up -d db
$DC run --rm --no-deps dockerize -wait tcp://db:5432 -timeout 40s

for cmd in keyper-0 keyper-1 keyper-2; do
    mkdb $cmd &
done

wait

$DC stop db
