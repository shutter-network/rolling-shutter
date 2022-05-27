import "../derived/wasm_exec";

let wasm = require("../derived/shutter-crypto.wasm");

const g = global || window || this || self;
if (typeof g.__wasm_functions__ === "undefined") {
  g.__wasm_functions__ = {};
}

function init() {
  let shutterCrypto;
  const go = new Go();
  if ("instantiateStreaming" in WebAssembly) {
    return WebAssembly.instantiateStreaming(fetch(wasm), go.importObject).then(
      (obj) => {
        shutterCrypto = obj.instance;
        go.run(shutterCrypto);
      }
    );
  } else {
    return fetch(wasm)
      .then((resp) => resp.arrayBuffer())
      .then((bytes) =>
        WebAssembly.instantiate(bytes, go.importObject).then((obj) => {
          shutterCrypto = obj.instance;
          go.run(shutterCrypto);
        })
      );
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
