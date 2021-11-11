const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var decryptorAddrs = await hre.getDecryptorAddresses();
  const decryptors = await ethers.getContract("Decryptors");
  const index = await decryptors.count();
  await decryptors.add(decryptorAddrs);
  await decryptors.append();

  const cfg = await ethers.getContract("DecryptorConfig");
  const currentBlock = await ethers.provider.getBlockNumber();

  await cfg.addNewCfg({
    activationBlockNumber: currentBlock + 10,
    setIndex: index,
  });
  console.log(
    "activationBlockNumber %s decryptors: %s",
    currentBlock + 10,
    decryptorAddrs
  );
};
