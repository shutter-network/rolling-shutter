## rolling-shutter gnosiskeyper

Run a Shutter keyper for Gnosis Chain

### Synopsis

This command runs a keyper node. It will connect to both a Gnosis and a
Shuttermint node which have to be started separately in advance.

```
rolling-shutter gnosiskeyper [flags]
```

### Options

```
      --config string   config file
  -h, --help            help for gnosiskeyper
```

### Options inherited from parent commands

```
      --logformat string   set log format, possible values:  min, short, long, max (default "long")
      --loglevel string    set log level, possible values:  warn, info, debug (default "info")
      --no-color           do not write colored logs
```

### SEE ALSO

* [rolling-shutter](rolling-shutter.md)	 - A collection of commands to run and interact with Rolling Shutter nodes
* [rolling-shutter gnosiskeyper dump-config](rolling-shutter_gnosiskeyper_dump-config.md)	 - Dump a 'gnosiskeyper' configuration file, based on given config and env vars
* [rolling-shutter gnosiskeyper generate-config](rolling-shutter_gnosiskeyper_generate-config.md)	 - Generate a 'gnosiskeyper' configuration file
* [rolling-shutter gnosiskeyper initdb](rolling-shutter_gnosiskeyper_initdb.md)	 - Initialize the database of the 'gnosiskeyper'

