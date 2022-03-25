const { ethers } = require("hardhat");
const { BigNumber } = require("ethers");

module.exports = async function (hre) {
  const fundValue = hre.deployConf.fundValue;
  if (fundValue == "") {
    console.log("fund: not doing any funding");
    return;
  }

  const [owner] = await ethers.getSigners();
  const keypers = await hre.getKeyperAddresses();
  const addresses = keypers;
  const value = ethers.utils.parseEther(fundValue);
  console.log(
    "fund: funding %s adresses with %s eth",
    addresses.length,
    fundValue
  );

  const txs = [];
  for (const a of addresses) {
    const balance = await ethers.provider.getBalance(a);
    const weiFund = ethers.utils.parseEther(fundValue);
    if (balance.gt(BigNumber.from(weiFund))) {
      console.log(a + " already funded");
      continue;
    }
    const tx = await owner.sendTransaction({
      to: a,
      value: value,
    });
    txs.push(tx);
  }
  for (const tx of txs) {
    await tx.wait();
  }
};
