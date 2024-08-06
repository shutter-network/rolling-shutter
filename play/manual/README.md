# Manual Setup

This directory contains a few scripts and corresponding config file templates to
semi-manually setup and run a set of keypers locally. Basic instructions:

0. Install requirements with `make install-asdf-plugins` and
   `make install-asdf`.
1. Build `rolling-shutter` with `make -C ../../rolling-shutter build`.
2. Run a postgres database, e.g. with `initdb -D work/db/` and
   `process-compose up` in the `/play` directory.
3. Set the variable `private_key` in `ethereum.sh` to a private key with funds
   on Chiado.
4. Clone the repos
   [gnosh-contracts](https://github.com/shutter-network/gnosh-contracts) and
   [shop-contracts](https://github.com/shutter-network/shop-contracts) in a
   common directory, including their submodules
   (`git submodule update --init --recursive`). Specify the relative path to
   this directory in the variable `contracts_root_dir`.
5. Run `./ethereum.sh`. This will output a set of contract addresses.
6. Copy and paste the contract addresses into the configuration files
   `keyper-0.toml`, `keyper-1.toml`, and `keyper-2.toml` under the section
   `[Gnosis.Contracts]`. Additionally, copy the keyper set manager address into
   `op-bootstrap-config.toml` under `KeyperSetManager`.

> Tip: A newly deployed validator registry won't have any registrations, so
> keypers won't produce any keys. Oftentimes it is therefore helpful to use the
> official validator registry on Chiado instead.

7. Run `./initdb.sh`
8. Run `./p2p.sh` and keep it running.
9. Run `./shuttermint.sh` and keep it running.
10. Wait until the activation block of the first keyper set is reached. The time
    it takes is defined by `activation_delta` defined in `ethereum.sh`.
11. Run the keypers with `./k.sh 0`, `./k.sh 1`, and `./k.sh 2`. Optionally, set
    the `ROLLING_SHUTTER_LOGLEVEL` environment variable to
    `:debug,dht:info,pubsub:info,swarm:info` to get reasonable log outputs.

If everything was successful, the keypers should start immediately with eon key
generation and once this succeeded decryption keys should be produced for every
Chiado slot with registered validator. Helpful log messages are:

- `"DKG process succeeded"`
- `"DKG process failed"`
- `"skipping slot as proposer is not registered"`
- `"sending decryption trigger"`.
