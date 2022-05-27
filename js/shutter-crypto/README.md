# shutter-crypto

This is a NPM package which provides the core crypto primitives necessary to interact with the Shutter Network.


## Installation

```
npm install @shutter-network/shutter-crypto@beta
```

## Usage

The module provides the following public functions:

### `shutterCrypto.init()`

Load and initialize the Go wasm library.
This Promise needs to be consumed before any other function in the library is called.

### `shutterCrypto.encrypt(message, eonPublicKey, epochId, sigma)`

...

### `shutterCrypto.decrypt(encryptedMessage, decryptionKey)`

...

### `shutterCrypto.verifyDecryptionKey(decryptionKey, eonPublicKey, epochId)`

...
