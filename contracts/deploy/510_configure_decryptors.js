const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var decryptorAddrs = await hre.getDecryptorAddresses();
  const decryptors = await ethers.getContract("Decryptors");
  const index = await decryptors.count();
  await decryptors.add(decryptorAddrs);
  await decryptors.append();

  const cfg = await ethers.getContract("DecryptorConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlock = currentBlock + 10;

  await cfg.addNewCfg({
    activationBlockNumber: activationBlock,
    setIndex: index,
  });
  console.log(
    "configure decryptors: activationBlockNumber %s decryptors: %s",
    activationBlock,
    decryptorAddrs
  );
};
