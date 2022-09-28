const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var decryptorAddrs = await hre.getDecryptorAddresses();
  const decryptors = await ethers.getContract("Decryptors");
  const index = await decryptors.count();
  let tx = await decryptors.add(decryptorAddrs);
  await tx.wait();
  tx = await decryptors.append();
  await tx.wait();

  const cfg = await ethers.getContract("DecryptorConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlock = currentBlock + 10;

  tx = await cfg.addNewCfg({
    activationBlockNumber: activationBlock,
    setIndex: index,
  });
  await tx.wait();
  console.log(
    "configure decryptors: activationBlockNumber %s decryptors: %s",
    activationBlock,
    decryptorAddrs
  );
};
