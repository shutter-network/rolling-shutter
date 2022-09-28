module.exports = {
  configure_keypers: configure_keypers,
};

const { ethers } = require("hardhat");

// const { inspect } = require("util");

async function configure_keypers(keyperAddrs) {
  const keypers = await ethers.getContract("Keypers");

  const index = (await keypers.count()) - 1;
  if (index >= 0) {
    const countAtIndex = (await keypers.countNth(index)) - 1;
    console.log("index:", index, "countAtIndex", countAtIndex);
    const currentAddrs = [];
    for (let i = 0; i < countAtIndex; i++) {
      currentAddrs.push(await keypers.at(index, i));
    }
    console.log(currentAddrs, keyperAddrs);
    if (currentAddrs === keyperAddrs) {
      console.log("Old and new keypres identical, not setting new config");
      return;
    }
  }

  await (await keypers.add(keyperAddrs)).wait();
  await (await keypers.append()).wait();

  const cfg = await ethers.getContract("KeyperConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlockNumber = currentBlock + 10;

  const tx = await cfg.addNewCfg({
    activationBlockNumber: activationBlockNumber,
    setIndex: index + 1,
    threshold: Math.ceil((keyperAddrs.length / 3) * 2),
  });
  await tx.wait();
  console.log(
    "configure keypers: activationBlockNumber %s, setIndex: %d, keypers: %s",
    activationBlockNumber,
    index + 1,
    keyperAddrs
  );
}
