const { configure_keypers } = require("../lib/configure-keypers.js");

module.exports = async function (hre) {
  var keyperAddrs = await hre.getKeyperAddresses();
  await configure_keypers(keyperAddrs, hre.deployConf.activationBlockOffset);
};
