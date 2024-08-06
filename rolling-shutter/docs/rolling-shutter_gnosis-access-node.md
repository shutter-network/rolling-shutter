## rolling-shutter gnosis-access-node

Run an access node for the keyper network of Shutterized Gnosis Chain

### Synopsis

This command runs a node that only relays messages, but doesn't create any on
its own. It is intended to be a stable node to connect to to receive messages.

```
rolling-shutter gnosis-access-node [flags]
```

### Options

```
      --config string   config file
  -h, --help            help for gnosis-access-node
```

### Options inherited from parent commands

```
      --logformat string   set log format, possible values:  min, short, long, max (default "long")
      --loglevel string    set log level, possible values:  warn, info, debug (default "info")
      --no-color           do not write colored logs
```

### SEE ALSO

* [rolling-shutter](rolling-shutter.md)	 - A collection of commands to run and interact with Rolling Shutter nodes
* [rolling-shutter gnosis-access-node dump-config](rolling-shutter_gnosis-access-node_dump-config.md)	 - Dump a 'gnosis-access-node' configuration file, based on given config and env vars
* [rolling-shutter gnosis-access-node generate-config](rolling-shutter_gnosis-access-node_generate-config.md)	 - Generate a 'gnosis-access-node' configuration file

