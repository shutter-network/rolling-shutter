module.exports = async function (hre) {
  const batchCounter = await hre.ethers.getContract("BatchCounter");
  const { deployer } = await hre.getNamedAccounts();
  const deployerSigner = await hre.ethers.getSigner(deployer);

  // Set the owner of batch counter to a special address so that it can only be controlled by the
  // state transition function.
  const currentOwner = batchCounter.owner();
  const ffAddress = hre.ethers.utils.getAddress(
    "0xffffffffffffffffffffffffffffffffffffffff"
  );
  if (currentOwner != ffAddress) {
    await batchCounter.connect(deployerSigner).transferOwnership(ffAddress);
  }
};
