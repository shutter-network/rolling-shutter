## rolling-shutter crypto aggregate

Aggregate key shares to construct a decryption key

### Synopsis

Aggregate key shares to construct a decryption key.

Pass the shares as the first and only positional argument in the form of hex
values separated by commas. The shares must be ordered by keyper index and
missing shares must be denoted by empty strings. Exactly "threshold" shares
must be provided.

Example: rolling-shutter crypto aggregate -t 2 0BC5CDC5778D473B881E73297AFB830
1D35830786C6A80CD289672536655470A0149BA7394DF240C96F7D60BAF94D0FD2A39B4314088E
AF94E3D1EB52106E718,,03646AE08A8EF00D0AE04294529466C0F7AC65C4D9B0ADEAD1461964A
6F784202B61B2EBD5DE8B80E787E9FD4DE4899880C2263B67EC478D88D3558B0C22DA66

```
rolling-shutter crypto aggregate [flags]
```

### Options

```
  -h, --help             help for aggregate
  -t, --threshold uint   threshold parameter
```

### Options inherited from parent commands

```
      --logformat string   set log format, possible values:  min, short, long, max (default "long")
      --loglevel string    set log level, possible values:  warn, info, debug (default "info")
      --no-color           do not write colored logs
```

### SEE ALSO

* [rolling-shutter crypto](rolling-shutter_crypto.md)	 - CLI tool to access crypto functions

