const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  await deployments.deploy("Collator", {
    contract: "AddrsSeq",
    from: deployer,
    args: [],
    log: true,
  });
  const c = await ethers.getContract("Collator");
  const tx = await c.append();
  await tx.wait();
};
