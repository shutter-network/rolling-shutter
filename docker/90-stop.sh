#!/usr/bin/env bash

source ./common.sh

echo "Stopping entire system"
$DC --profile dev down
