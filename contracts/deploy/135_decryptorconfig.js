const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  const decryptors = await ethers.getContract("Decryptors");
  const registry = await ethers.getContract("BLSRegistry");

  await deployments.deploy("DecryptorConfig", {
    contract: "DecryptorsConfigsList",
    from: deployer,
    args: [decryptors.address, registry.address],
    log: true,
  });
};
