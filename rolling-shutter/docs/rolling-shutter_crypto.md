## rolling-shutter crypto

CLI tool to access crypto functions

### Synopsis

This command provides utility functions to manually encrypt messages with an eon
key, decrypt them with a decryption key, and check that a decryption key is correct.

### Options

```
  -h, --help   help for crypto
```

### Options inherited from parent commands

```
      --logformat string   set log format, possible values:  min, short, long, max (default "long")
      --loglevel string    set log level, possible values:  warn, info, debug (default "info")
      --no-color           do not write colored logs
```

### SEE ALSO

* [rolling-shutter](rolling-shutter.md)	 - A collection of commands to run and interact with Rolling Shutter nodes
* [rolling-shutter crypto aggregate](rolling-shutter_crypto_aggregate.md)	 - Aggregate key shares to construct a decryption key
* [rolling-shutter crypto decrypt](rolling-shutter_crypto_decrypt.md)	 - Decrypt the message given as positional argument
* [rolling-shutter crypto encrypt](rolling-shutter_crypto_encrypt.md)	 - Encrypt the message given as positional argument
* [rolling-shutter crypto verify-key](rolling-shutter_crypto_verify-key.md)	 - Check that the decryption key given as positional argument is correct

