## rolling-shutter primevkeyper

Run a Shutter keyper for PrimeV POC

### Synopsis

This command runs a keyper node. It will connect to both a PrimeV and a
Shuttermint node which have to be started separately in advance.

```
rolling-shutter primevkeyper [flags]
```

### Options

```
      --config string   config file
  -h, --help            help for primevkeyper
```

### Options inherited from parent commands

```
      --logformat string   set log format, possible values:  min, short, long, max (default "long")
      --loglevel string    set log level, possible values:  warn, info, debug (default "info")
      --no-color           do not write colored logs
```

### SEE ALSO

* [rolling-shutter](rolling-shutter.md)	 - A collection of commands to run and interact with Rolling Shutter nodes
* [rolling-shutter primevkeyper dump-config](rolling-shutter_primevkeyper_dump-config.md)	 - Dump a 'primevkeyper' configuration file, based on given config and env vars
* [rolling-shutter primevkeyper generate-config](rolling-shutter_primevkeyper_generate-config.md)	 - Generate a 'primevkeyper' configuration file
* [rolling-shutter primevkeyper initdb](rolling-shutter_primevkeyper_initdb.md)	 - Initialize the database of the 'primevkeyper'

