module.exports = {
  configure_keypers: configure_keypers,
};

const { ethers } = require("hardhat");

// const { inspect } = require("util");

async function configure_keypers(keyperAddrs) {
  const keypers = await ethers.getContract("Keypers");

  const index = await keypers.count();
  await keypers.add(keyperAddrs);
  await keypers.append();

  const cfg = await ethers.getContract("KeyperConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlockNumber = currentBlock + 10;

  await cfg.addNewCfg({
    activationBlockNumber: activationBlockNumber,
    setIndex: index,
    threshold: Math.ceil((keyperAddrs.length / 3) * 2),
  });
  console.log(
    "configure keypers: activationBlockNumber %s, setIndex: %d, keypers: %s",
    activationBlockNumber,
    index,
    keyperAddrs
  );
}
