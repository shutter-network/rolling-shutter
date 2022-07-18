# shutter-crypto

This is a NPM package which provides the core crypto primitives necessary to
interact with the Shutter Network.

## Installation

```
npm install @shutter-network/shutter-crypto@beta
```

## Usage

The module provides the following public functions:

### `async shutterCrypto.init(wasmUrlOrPath)`

Load and initialize the Go wasm library. This Promise needs to be consumed
before any other function in the library is called.

On Node the `wasmUrlOrPath` parameter is optional. If not given it will be
determined automatically.

In a Web context the path to the `shutter-crypto.wasm` file needs to be given
(since it appears no standard cross framework way of automatically determining a
path is available).

### `async shutterCrypto.encrypt(message, eonPublicKey, epochId, sigma)`

...

### `async shutterCrypto.decrypt(encryptedMessage, decryptionKey)`

...

### `async shutterCrypto.verifyDecryptionKey(decryptionKey, eonPublicKey, epochId)`

...
