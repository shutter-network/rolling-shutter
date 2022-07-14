import "../derived/wasm_exec";
import { isBrowser, isNode } from "browser-or-node";

const g = global || window || this || self;
if (typeof g.__wasm_functions__ === "undefined") {
  g.__wasm_functions__ = {};
}

function init(wasmUrlOrPath) {
  let shutterCrypto;
  const go = new Go(); // eslint-disable-line no-undef
  if (isBrowser) {
    if ("instantiateStreaming" in WebAssembly) {
      return WebAssembly.instantiateStreaming(
        fetch(wasmUrlOrPath),
        go.importObject
      ).then((obj) => {
        shutterCrypto = obj.instance;
        go.run(shutterCrypto);
      });
    } else {
      return fetch(wasmUrlOrPath)
        .then((resp) => resp.arrayBuffer())
        .then((bytes) =>
          WebAssembly.instantiate(bytes, go.importObject).then((obj) => {
            shutterCrypto = obj.instance;
            go.run(shutterCrypto);
          })
        );
    }
  } else if (isNode) {
    const fs = __non_webpack_require__("fs"); // eslint-disable-line no-undef
    WebAssembly.instantiate(
      fs.readFileSync(wasmUrlOrPath),
      go.importObject
    ).then((obj) => {
      shutterCrypto = obj.instance;
      go.run(shutterCrypto);
    });
  } else {
    throw "Neither Browser nor Node; not supported.";
  }
}

function _checkInitialized() {
  if (typeof g.__wasm_functions__.encrypt === "undefined") {
    throw "You need to consume the 'shutterCrypto.init()' promise before using the module functions.";
  }
}

function encrypt(message, eonPublicKey, epochId, sigma) {
  _checkInitialized();
  return g.__wasm_functions__.encrypt(message, eonPublicKey, epochId, sigma);
}

function decrypt(encryptedMessage, decryptionKey) {
  _checkInitialized();
  return g.__wasm_functions__.decrypt(encryptedMessage, decryptionKey);
}

function verifyDecryptionKey(decryptionKey, eonPublicKey, epochId) {
  _checkInitialized();
  return g.__wasm_functions__.verifyDecryptionKey(
    decryptionKey,
    eonPublicKey,
    epochId
  );
}

export { init, encrypt, decrypt, verifyDecryptionKey };
