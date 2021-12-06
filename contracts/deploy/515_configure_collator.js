const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var collatorAddress = await hre.getCollatorAddress();
  const collator = await ethers.getContract("Collator");
  const index = await collator.count();
  await collator.add([collatorAddress]);
  await collator.append();

  const cfg = await ethers.getContract("CollatorConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlock = currentBlock + 10;

  await cfg.addNewCfg({
    activationBlockNumber: activationBlock,
    setIndex: index,
  });
  console.log(
    "configure collator: activationBlockNumber %s collator: %s",
    activationBlock,
    collatorAddress
  );
};
