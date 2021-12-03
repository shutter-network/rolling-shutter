const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();

  const decryptors = await ethers.getContract("Decryptors");

  await deployments.deploy("BLSRegistry", {
    contract: "Registry",
    from: deployer,
    args: [decryptors.address],
    log: true,
  });
};
