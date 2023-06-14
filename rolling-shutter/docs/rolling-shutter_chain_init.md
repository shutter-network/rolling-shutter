## rolling-shutter chain init

Create a config file for a Shuttermint node

```
rolling-shutter chain init [flags]
```

### Options

```
      --blocktime float          block time in seconds (default 1)
      --dev                      turn on devmode (disables validator set changes)
      --genesis-keyper strings   genesis keyper address
  -h, --help                     help for init
      --index int                keyper index
      --initial-eon uint         initial eon
      --listen-address string    tendermint RPC listen address (default "tcp://127.0.0.1:26657")
      --role string              tendermint node role (validator, sentry, seed) (default "validator")
      --root string              root directory
```

### Options inherited from parent commands

```
      --logformat string   set log format, possible values:  min, short, long, max (default "long")
      --loglevel string    set log level, possible values:  warn, info, debug (default "info")
      --no-color           do not write colored logs
```

### SEE ALSO

* [rolling-shutter chain](rolling-shutter_chain.md)	 - Run a node for Shutter's Tendermint chain

