const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  var collator = await ethers.getContract("Collator");
  await deployments.deploy("CollatorConfig", {
    contract: "CollatorConfigsList",
    from: deployer,
    args: [collator.address],
    log: true,
  });
};
