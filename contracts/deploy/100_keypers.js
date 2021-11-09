const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  await deployments.deploy("Keypers", {
    contract: "AddrsSeq",
    from: deployer,
    args: [],
    log: true,
  });
  const c = await ethers.getContract("Keypers");
  await c.append();
};
