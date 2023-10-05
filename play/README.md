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
