const { isNode } = require("browser-or-node");
const crypto = isNode ? __non_webpack_require__("crypto") : window.crypto; // eslint-disable-line no-undef
module.exports = crypto;
