const { ethers } = require("hardhat");
const deployOptions = require("../lib/deploy_options.js");
const waitForDeployment = require("../lib/wait_for_deployment.js");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();

  const decryptors = await ethers.getContract("Decryptors");

  const deployment = await deployments.deploy(
    "BLSRegistry",
    Object.assign(deployOptions, {
      contract: "Registry",
      from: deployer,
      args: [decryptors.address],
    })
  );
  await waitForDeployment(deployment);
};
