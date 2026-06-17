#!/usr/bin/env bash

source ./common.sh

echo "Submitting bootstrap transaction"

docker run --rm -it --network snapshutter_default -v /root/rolling-shutter/docker/config:/config -v /root/rolling-shutter/docker/data/bootstrap:/data rolling-shutter op-bootstrap --config /config/bootstrap.toml
