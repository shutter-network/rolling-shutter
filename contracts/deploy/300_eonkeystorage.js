const deployOptions = require("../lib/deploy_options.js");
const waitForDeployment = require("../lib/wait_for_deployment.js");

module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  const deployment = await deployments.deploy(
    "EonKeyStorage",
    Object.assign(deployOptions, {
      contract: "EonKeyStorage",
      from: deployer,
      args: [],
    })
  );
  await waitForDeployment(deployment);
};
