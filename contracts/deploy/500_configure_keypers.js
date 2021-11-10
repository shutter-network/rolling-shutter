const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var keyperAddrs = await hre.getKeyperAddresses();
  const keypers = await ethers.getContract("Keypers");
  const index = await keypers.count();
  await keypers.add(keyperAddrs);
  await keypers.append();

  const cfg = await ethers.getContract("KeyperConfig");
  const currentBlock = await ethers.provider.getBlockNumber();

  await cfg.addNewCfg({
    activationBlockNumber: currentBlock + 10,
    setIndex: index,
  });
  console.log(
    "activationBlockNumber %s keypers: %s",
    currentBlock + 10,
    keyperAddrs
  );
};
