## rolling-shutter shutterservicekeyper

Run a Shutter keyper for Shutter Service

### Synopsis

This command runs a keyper node. It will connect to both a Shutter service and a
Shuttermint node which have to be started separately in advance.

```
rolling-shutter shutterservicekeyper [flags]
```

### Options

```
      --config string   config file
  -h, --help            help for shutterservicekeyper
```

### Options inherited from parent commands

```
      --logformat string   set log format, possible values:  min, short, long, max (default "long")
      --loglevel string    set log level, possible values:  warn, info, debug (default "info")
      --no-color           do not write colored logs
```

### SEE ALSO

* [rolling-shutter](rolling-shutter.md)	 - A collection of commands to run and interact with Rolling Shutter nodes
* [rolling-shutter shutterservicekeyper dump-config](rolling-shutter_shutterservicekeyper_dump-config.md)	 - Dump a 'shutterservicekeyper' configuration file, based on given config and env vars
* [rolling-shutter shutterservicekeyper generate-config](rolling-shutter_shutterservicekeyper_generate-config.md)	 - Generate a 'shutterservicekeyper' configuration file
* [rolling-shutter shutterservicekeyper initdb](rolling-shutter_shutterservicekeyper_initdb.md)	 - Initialize the database of the 'shutterservicekeyper'

