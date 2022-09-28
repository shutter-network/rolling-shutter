const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var collatorAddress = await hre.getCollatorAddress();
  const collator = await ethers.getContract("Collator");
  const index = await collator.count();
  let tx = await collator.add([collatorAddress]);
  await tx.wait();
  tx = await collator.append();
  await tx.wait();

  const cfg = await ethers.getContract("CollatorConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlock = currentBlock + 10;

  tx = await cfg.addNewCfg({
    activationBlockNumber: activationBlock,
    setIndex: index,
  });
  await tx.wait();
  console.log(
    "configure collator: activationBlockNumber %s collator: %s",
    activationBlock,
    collatorAddress
  );
};
