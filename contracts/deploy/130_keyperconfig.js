const { ethers } = require("hardhat");
const deployOptions = require("../lib/deploy_options.js");
const waitForDeployment = require("../lib/wait_for_deployment.js");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  var keypers = await ethers.getContract("Keypers");
  const deployment = await deployments.deploy(
    "KeyperConfig",
    Object.assign(deployOptions, {
      contract: "KeypersConfigsList",
      from: deployer,
      args: [keypers.address],
    })
  );
  await waitForDeployment(deployment);
};
