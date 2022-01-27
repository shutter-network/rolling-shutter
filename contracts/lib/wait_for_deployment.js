const hre = require("hardhat");

// wait for the configured number of confirmations for a contract deployment. hardhat-deploy's
// waitConfirmation option should do this automatically, but it's broken:
// https://github.com/wighawag/hardhat-deploy/issues/267
module.exports = async (deployment) => {
  const deployTx = await hre.ethers.provider.getTransaction(
    deployment.transactionHash
  );
  await deployTx.wait(hre.numConfirmations);
};
