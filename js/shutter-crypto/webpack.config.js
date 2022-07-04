const path = require("path");

/* global __dirname */

module.exports = {
  mode: "production",
  performance: {
    maxAssetSize: 350_000,
  },
  output: {
    path: path.resolve(__dirname, "dist"),
    filename: "shutter-crypto.js",
    publicPath: "",
    globalObject: "this",
    library: {
      name: "shutterCrypto",
      type: "umd",
    },
  },
  module: {
    noParse: /wasm_exec.js/,
    rules: [{ test: /\.wasm$/, type: "asset" }],
  },
  node: {
    // Disable mangling node's `__dirname` property since we need it to load the WASM file
    __dirname: false,
  },
};
