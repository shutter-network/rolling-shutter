import "../derived/wasm_exec";
import { isBrowser, isNode } from "browser-or-node";

const g = global || window || this || self;
if (typeof g.__wasm_functions__ === "undefined") {
  g.__wasm_functions__ = {};
}

async function init(wasmUrlOrPath) {
  let shutterCrypto;
  const go = new Go(); // eslint-disable-line no-undef
  if (isBrowser) {
    if ("instantiateStreaming" in WebAssembly) {
      const obj = await WebAssembly.instantiateStreaming(
        fetch(wasmUrlOrPath),
        go.importObject
      );
      shutterCrypto = obj.instance;
      go.run(shutterCrypto);
    } else {
      const resp = await fetch(wasmUrlOrPath);
      const bytes = await resp.arrayBuffer();
      const obj = WebAssembly.instantiate(bytes, go.importObject);
      shutterCrypto = obj.instance;
      go.run(shutterCrypto);
    }
  } else if (isNode) {
    const fs = __non_webpack_require__("fs"); // eslint-disable-line no-undef
    const obj = await WebAssembly.instantiate(
      fs.readFileSync(wasmUrlOrPath),
      go.importObject
    );
    shutterCrypto = obj.instance;
    go.run(shutterCrypto);
  } else {
    throw "Neither Browser nor Node; not supported.";
  }
}

function _checkInitialized() {
  if (typeof g.__wasm_functions__.encrypt === "undefined") {
    throw "You need to consume the 'shutterCrypto.init()' promise before using the module functions.";
  }
}

function _throwOnError(result) {
  if (result.startsWith("Error:")) {
    throw result;
  }
}

function _hexToUint8Array(hex) {
  if (hex.startsWith("0x")) {
    hex = hex.slice(2);
  }
  if (hex.length % 2 != 0) {
    hex = "0" + hex;
  }
  let bytes = [];
  for (let i = 0; i < hex.length; i += 2) {
    bytes.push(parseInt(hex.substring(i, i + 2), 16));
  }
  return Uint8Array.from(bytes);
}

async function encrypt(message, eonPublicKey, epochId, sigma) {
  _checkInitialized();
  const result = await g.__wasm_functions__.encrypt(
    message,
    eonPublicKey,
    epochId,
    sigma
  );
  _throwOnError(result);
  return _hexToUint8Array(result);
}

async function decrypt(encryptedMessage, decryptionKey) {
  _checkInitialized();
  const result = await g.__wasm_functions__.decrypt(
    encryptedMessage,
    decryptionKey
  );
  _throwOnError(result);
  return _hexToUint8Array(result);
}

async function verifyDecryptionKey(decryptionKey, eonPublicKey, epochId) {
  _checkInitialized();
  const result = await g.__wasm_functions__.verifyDecryptionKey(
    decryptionKey,
    eonPublicKey,
    epochId
  );
  _throwOnError(result);
  return result;
}

export { init, encrypt, decrypt, verifyDecryptionKey };
