const { ethers } = require("hardhat");
const deployOptions = require("../lib/deploy_options.js");
const waitForDeployment = require("../lib/wait_for_deployment.js");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  const deployResult = await deployments.deploy(
    "Decryptors",
    Object.assign(deployOptions, {
      contract: "AddrsSeq",
      from: deployer,
      args: [],
    })
  );
  await waitForDeployment(deployResult);
  if (deployResult.newlyDeployed) {
    const c = await ethers.getContract("Decryptors");
    const tx = await c.append();
    await tx.wait(hre.numConfirmations);
  }
};
