const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { getNamedAccounts } = hre;
  const { keyper0, keyper1, keyper2 } = await getNamedAccounts();

  const keypers = await ethers.getContract("Keypers");
  const index = await keypers.count();
  await keypers.add([keyper0, keyper1, keyper2]);
  await keypers.append();

  const cfg = await ethers.getContract("KeyperConfig");
  const currentBlock = await ethers.provider.getBlockNumber();

  await cfg.addNewCfg({
    activationBlockNumber: currentBlock + 10,
    setIndex: index,
  });
};
