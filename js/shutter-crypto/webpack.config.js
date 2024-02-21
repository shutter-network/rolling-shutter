import { dirname, path } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));

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
      crypto: path.resolve(__dirname, "src", "_node18_crypto_fallback.js"),
      util: false,
    },
  },
};
