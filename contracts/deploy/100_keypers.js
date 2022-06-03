const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  const deployResult = await deployments.deploy("Keypers", {
    contract: "AddrsSeq",
    from: deployer,
    args: [],
    log: true,
  });
  if (deployResult.newlyDeployed) {
    const c = await ethers.getContract("Keypers");
    const tx = await c.append();
    await tx.wait();
  }
};
