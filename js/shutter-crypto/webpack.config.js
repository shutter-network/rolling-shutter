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
  resolve: {
    fallback: {
      fs: false,
      crypto: false,
      util: false,
    },
  },
};
