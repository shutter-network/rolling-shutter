# test setup using babashka

Download babashka, make sure bb is in your PATH, as well as rolling-shutter's
bin directory.

Set the environment variable `ROLLING_SHUTTER_ROOT` to the directory containing
the checkout of the rolling-shutter repository.

## shuttermint environment variables

Set the following environment variables (you're free to choose a different
values here) before running `bb init` or `bb boot`:

```
export ROLLING_SHUTTER_BOOTSTRAP_SIGNING_KEY=479968ffa5ee4c84514a477a8f15f3db0413964fd4c20b08a55fed9fed790fad
export ROLLING_SHUTTER_CHAIN_GENESIS_KEYPER=0x440Dc6F164e9241F04d282215ceF2780cd0B755e
```

## prepare psql (via docker)

```
docker run --rm -it -e POSTGRES_USER=$(whoami) -e POSTGRES_PASSWORD=password -e POSTGRES_DB=testdb -v /tmp:/tmp --net=host -v /tmp/projectdir:/home/circleci/project -v /tmp/datadir:/var/lib/postgresql/data cimg/postgres:13.9
```

## keyper only test setup

1. bb init; bb chain
2. bb node
3. bb boot
4. bb peer keyper-0.toml keyper-1.toml keyper-2.toml p2p-0.toml
5. bb p2p 0
6. bb k 0
7. bb k 1
8. bb k 2

## full test setup

1.  bb init; bb chain
2.  bb node
3.  bb peer collator.toml mock.toml keyper-0.toml keyper-1.toml keyper-2.toml
    p2p-0.toml; bb boot; bb sequencer
4.  bb p2p 0
5.  bb k 0
6.  bb k 1
7.  bb k 2
8.  bb collator

## Whole system tests

Make sure you have java jdk 17 as well as clojure installed.

Run `bb test-system`.

## nitro

Run `bb nitro` to start an arbitrum nitro test setup. This requires docker and
docker-compose to be installed.

## Running third party clients in a test

There is an example test task `keyper-dkg-external`, that shows how a third
party client can be addressed.

Instead of three `keyper` instances, it starts only two and one external
executable. That external executable can be defined with `BB_EXTERNAL_COMMAND`.
For convenience, there is also the optional `BB_EXTERNAL_WD`, to manipulate the
working directory of the external call.

As an example, here is how one of the three original test keypers can be called
manually:

```
BB_EXTERNAL_WD="$(pwd)" BB_EXTERNAL_COMMAND="../rolling-shutter/bin/rolling-shutter keyper --config work/keyper-dkg-external/keyper-2.toml" clojure -M:test keyper-dkg-external
```

(You can do the same inside a docker container (see below)).

### Parameters in the environment

You can find the test system parameters in the environment of the
`BB_EXTERNAL_COMMAND`, e.g.:

```
# run "/bin/sh -c env" as external command to dump the environment:
BB_EXTERNAL_COMMAND="/bin/sh -c env" clojure -M:test keyper-dkg-external
...
# find the values in the test log:
cat work/keyper-dkg-external/logs/keyper-external-*|grep KPR
KPR_P2P_PORT=23102
KPR_DKG_PHASE_LENGTH=8
KPR_ENVIRONMENT=local
KPR_ETHEREUM_URL=http://127.0.0.1:8545/
KPR_DKG_START_BLOCK_DELTA=5
KPR_HTTP_LISTEN_ADDRESS=:24003
KPR_CONTRACTS_URL=http://127.0.0.1:8545/
KPR_LISTEN_ADDRESSES=/ip4/127.0.0.1/tcp/23102
```

Here you see, that the keyper implementation under tests, should use port
`23102` for its `p2p` connection, the ethereum node is running at
`http://127.0.0.1:8545/` etc. Not all of these may be relevant to your
implementation.

### Docker image

There are ready to use docker images at
https://ghcr.io/shutter-network/babashka-play

To run a test inside a container, run:

```
docker run --platform=linux/amd64 --rm -it ghcr.io/shutter-network/babashka-play:latest

process-compose -t=false up & # to start postgres

# RUN TEST
#e.g.
bb test-system
```

Keep in mind that the configurations are generated within the container on first
invocation. This means that things like private keys and peer addresses are
changing when the container is started with the `docker run --rm` command. To
keep those values persistent, either the configuration files have to be copied
to the container, or the container should be reused throughout runs.

### Self contained testrunner

It is possible to compile the testrunner into a self contained `.jar` file, by
running

```
clojure -T:build
```

If you have such a `sht-standalone.jar` you can run the above test (without the
need to install clojure), by calling

```
BB_EXTERNAL_COMMAND="../rolling-shutter/bin/rolling-shutter keyper --config work/keyper-dkg-external/keyper-2.toml" java -jar sht-standalone.jar -M:test keyper-dkg-external
```

(The `circleci` integration tests make use of this.)
