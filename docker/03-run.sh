#!/usr/bin/env bash

source ./common.sh

echo "Starting entire system"
$DC --profile dev up -d
sleep 5
$DC ps
