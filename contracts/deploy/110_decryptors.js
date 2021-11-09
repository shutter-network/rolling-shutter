const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  await deployments.deploy("Decryptors", {
    contract: "AddrsSeq",
    from: deployer,
    args: [],
    log: true,
  });
  const c = await ethers.getContract("Decryptors");
  await c.append();
};
