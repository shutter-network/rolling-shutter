const { ethers } = require("hardhat");
const deployOptions = require("../lib/deploy_options.js");
const waitForDeployment = require("../lib/wait_for_deployment.js");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  var collator = await ethers.getContract("Collator");
  const deployment = await deployments.deploy(
    "CollatorConfig",
    Object.assign(deployOptions, {
      contract: "CollatorConfigsList",
      from: deployer,
      args: [collator.address],
    })
  );
  await waitForDeployment(deployment);
};
