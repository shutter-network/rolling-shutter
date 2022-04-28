module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  await deployments.deploy("BatchCounter", {
    contract: "BatchCounter",
    from: deployer,
    args: [],
    log: true,
  });
};
