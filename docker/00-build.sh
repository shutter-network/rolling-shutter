#!/usr/bin/env bash

source ./common.sh

$DC --profile dev down
$DC --profile dev build
