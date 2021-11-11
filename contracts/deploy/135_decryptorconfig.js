const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  const decryptors = await ethers.getContract("Decryptors");
  const pubkeys = await ethers.getContract("BLSPublicKeyRegistry");
  const signatures = await ethers.getContract("BLSSignatureRegistry");

  await deployments.deploy("DecryptorConfig", {
    contract: "DecryptorsConfigsList",
    from: deployer,
    args: [decryptors.address, pubkeys.address, signatures.address],
    log: true,
  });
};
