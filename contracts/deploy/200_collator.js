const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  const deployResult = await deployments.deploy("Collator", {
    contract: "AddrsSeq",
    from: deployer,
    args: [],
    log: true,
  });
  if (deployResult.newlyDeployed) {
    const c = await ethers.getContract("Collator");
    await c.append();
  }
};
