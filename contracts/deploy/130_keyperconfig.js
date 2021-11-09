const { ethers } = require("hardhat");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  var keypers = await ethers.getContract("Keypers");
  await deployments.deploy("KeyperConfig", {
    contract: "KeypersConfigsList",
    from: deployer,
    args: [keypers.address],
    log: true,
  });
};
