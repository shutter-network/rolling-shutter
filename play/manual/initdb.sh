#!/bin/bash

dropdb --if-exists keyper-db-0
dropdb --if-exists keyper-db-1
dropdb --if-exists keyper-db-2
createdb keyper-db-0
createdb keyper-db-1
createdb keyper-db-2
./rolling-shutter gnosiskeyper initdb --config keyper-0.toml
./rolling-shutter gnosiskeyper initdb --config keyper-1.toml
./rolling-shutter gnosiskeyper initdb --config keyper-2.toml
