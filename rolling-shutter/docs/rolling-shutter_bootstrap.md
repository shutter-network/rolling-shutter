## rolling-shutter bootstrap

Bootstrap Shuttermint by submitting the initial batch config

### Synopsis

This command sends a batch config to the Shuttermint chain in a message signed
with the given private key. This will instruct a newly created chain to update
its validator set according to the keyper set defined in the batch config. The
private key must correspond to the initial validator address as defined in the
chain's genesis config.

```
rolling-shutter bootstrap [flags]
```

### Options

```
      --deployment-dir string    Deployment directory (default "./deployments/localhost")
      --ethereum-url string      Ethereum URL (default "http://localhost:8545")
  -h, --help                     help for bootstrap
  -i, --index int                keyper config index to bootstrap with (use latest if negative) (default 1)
  -s, --shuttermint-url string   Shuttermint RPC URL (default "http://localhost:26657")
  -k, --signing-key string       private key of the keyper to send the message with
```

### Options inherited from parent commands

```
      --logformat string   set log format, possible values:  min, short, long, max (default "long")
      --loglevel string    set log level, possible values:  warn, info, debug (default "info")
      --no-color           do not write colored logs
```

### SEE ALSO

* [rolling-shutter](rolling-shutter.md)	 - A collection of commands to run and interact with Rolling Shutter nodes

