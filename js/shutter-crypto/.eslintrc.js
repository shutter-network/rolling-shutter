module.exports = {
  extends: "eslint:recommended",
  ignorePatterns: ["derived/*.js"],
  parserOptions: {
    sourceType: "module",
  },
  env: {
    browser: true,
    jest: true,
    node: true,
  },
};
