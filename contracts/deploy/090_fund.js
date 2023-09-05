const { ethers } = require("hardhat");
const { fund } = require("../lib/fund.js");

module.exports = async function (hre) {
  const fundValue = hre.deployConf.fundValue;
  const { bank, deployer } = await hre.getNamedAccounts();
  const bankSigner = await ethers.getSigner(bank);

  let addresses = [];
  if (deployer !== bank) {
    addresses.push(deployer);
  }
  addresses.push(...(await hre.getKeyperAddresses(0)));
  addresses.push(await hre.getCollatorAddress());
  await fund(addresses, bankSigner, fundValue);
};
