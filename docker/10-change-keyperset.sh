#!/usr/bin/env bash

source ./common.sh

$DC run --rm --no-deps deploy-contracts run scripts/change-keypers.js
