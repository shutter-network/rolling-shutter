#!/usr/bin/env bash

source ./common.sh

TARGET_INDEX=${1:-1}

$DC run --rm --no-deps -e KEYPER_SET_INDEX=${TARGET_INDEX} deploy-contracts run scripts/change-keypers.js
