const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var keyperAddrs = await hre.getKeyperAddresses();
  const keypers = await ethers.getContract("Keypers");
  const index = await keypers.count();
  await keypers.add(keyperAddrs);
  await keypers.append();

  const cfg = await ethers.getContract("KeyperConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlockNumber = currentBlock + 10;

  await cfg.addNewCfg({
    activationBlockNumber: activationBlockNumber,
    setIndex: index,
  });
  console.log(
    "configure keypers: activationBlockNumber %s keypers: %s",
    activationBlockNumber,
    keyperAddrs
  );
};
