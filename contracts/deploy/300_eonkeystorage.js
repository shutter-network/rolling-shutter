module.exports = async function (hre) {
  const { deployments, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  await deployments.deploy("EonKeyStorage", {
    contract: "EonKeyStorage",
    from: deployer,
    args: [],
    log: true,
  });
};
