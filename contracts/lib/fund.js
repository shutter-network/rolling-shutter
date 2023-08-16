module.exports = {
  fund: fund,
};

const { ethers } = require("hardhat");
const { BigNumber } = require("ethers");

async function fund(addresses, bankSigner, fundValue = "1000") {
  if (fundValue == "") {
    // TODO return error?
    console.log("fund: not doing any funding");
    return;
  }

  // TODO errorhandling?
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
    const tx = await bankSigner.sendTransaction({
      to: a,
      value: value,
    });
    txs.push(tx);
  }
  const txPromises = [];
  for (const tx of txs) {
    txPromises.push(tx.wait());
  }
  return Promise.all(txPromises);
}
