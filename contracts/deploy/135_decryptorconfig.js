const { ethers } = require("hardhat");
const deployOptions = require("../lib/deploy_options.js");
const waitForDeployment = require("../lib/wait_for_deployment.js");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  const decryptors = await ethers.getContract("Decryptors");
  const registry = await ethers.getContract("BLSRegistry");

  const deployment = await deployments.deploy(
    "DecryptorConfig",
    Object.assign(deployOptions, {
      contract: "DecryptorsConfigsList",
      from: deployer,
      args: [decryptors.address, registry.address],
    })
  );
  await waitForDeployment(deployment);
};
