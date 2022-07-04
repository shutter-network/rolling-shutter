import "../derived/wasm_exec";
import { isBrowser, isNode } from "browser-or-node";

let wasm = require("../derived/shutter-crypto.wasm");

const g = global || window || this || self;
if (typeof g.__wasm_functions__ === "undefined") {
  g.__wasm_functions__ = {};
}

/* global Go, __non_webpack_require__, __dirname, __webpack_require__ */

/* Slightly modified copy from webpack/runtime/publicPath
 * Webpack's automatic publicPath doesn't work in node, so we need to manually
 * handle it to be compatible with both.
 */
function getScriptUrl() {
  var scriptUrl;
  if (__webpack_require__.g.importScripts)
    scriptUrl = __webpack_require__.g.location + "";
  var document = __webpack_require__.g.document;
  if (!scriptUrl && document) {
    if (document.currentScript) scriptUrl = document.currentScript.src;
    if (!scriptUrl) {
      var scripts = document.getElementsByTagName("script");
      if (scripts.length) scriptUrl = scripts[scripts.length - 1].src;
    }
  }
  // When supporting browsers where an automatic publicPath is not supported you must specify an output.publicPath manually via configuration
  // or pass an empty string ("") and set the __webpack_public_path__ variable from your code to use your own logic.
  if (!scriptUrl)
    throw new Error("Automatic publicPath is not supported in this browser");
  scriptUrl = scriptUrl
    .replace(/#.*$/, "")
    .replace(/\?.*$/, "")
    .replace(/\/[^/]+$/, "/");
  return scriptUrl;
}

function init() {
  let shutterCrypto;
  const go = new Go();
  if (isBrowser) {
    wasm = getScriptUrl() + wasm;
    if ("instantiateStreaming" in WebAssembly) {
      return WebAssembly.instantiateStreaming(
        fetch(wasm),
        go.importObject
      ).then((obj) => {
        shutterCrypto = obj.instance;
        go.run(shutterCrypto);
      });
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
  } else if (isNode) {
    const fs = __non_webpack_require__("fs");
    const path = __non_webpack_require__("path");
    WebAssembly.instantiate(
      fs.readFileSync(path.join(__dirname, wasm)),
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
