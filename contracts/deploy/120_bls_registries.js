const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();

  const decryptors = await ethers.getContract("Keypers");

  await deployments.deploy("BLSPublicKeyRegistry", {
    contract: "Registry",
    from: deployer,
    args: [decryptors.address],
    log: true,
  });
  await deployments.deploy("BLSSignatureRegistry", {
    contract: "Registry",
    from: deployer,
    args: [decryptors.address],
    log: true,
  });
};
