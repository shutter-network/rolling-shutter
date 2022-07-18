const path = require("path");

module.exports = {
  mode: "production",
  output: {
    path: path.resolve(__dirname, "dist"),
    filename: "shutter-crypto.js",
    globalObject: "this",
    library: {
      name: "shutterCrypto",
      type: "umd",
    },
  },
  node: {
    // Disable mangling node's `__dirname` property since we need it to load the WASM file
    __dirname: false,
  },
  resolve: {
    fallback: {
      fs: false,
      crypto: false,
      util: false,
    },
  },
};
