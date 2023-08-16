module.exports = {
  configure_keypers: configure_keypers,
};

const { ethers } = require("hardhat");

// const { inspect } = require("util");

//TODO since we want to call this for new keypers
// this should also check and eventually fund the
// keypers if they are below the target funding.
async function configure_keypers(keyperAddrs, blockOffset = 10) {
  if (keyperAddrs.length == 0) {
    console.log("WARNING: cannot configure keypers: no keyper addresses given");
    return;
  }

  const keypers = await ethers.getContract("Keypers");
  const lastSetIndex = (await keypers.count()) - 1;
  let configSetIndex;

  let keyperSetAdded = true;
  if (
    (await keypers.countNth(lastSetIndex)).toNumber() !== keyperAddrs.length
  ) {
    keyperSetAdded = false;
  } else {
    for (const i of Array(keyperAddrs.length).keys()) {
      if ((await keypers.at(lastSetIndex, i)) !== keyperAddrs[i]) {
        keyperSetAdded = false;
        break;
      }
    }
  }
  if (keyperSetAdded) {
    console.log("Keyper set already added");
    configSetIndex = lastSetIndex;
  } else {
    console.log(keyperAddrs);
    const tx = await keypers.add(keyperAddrs);
    await tx.wait();
    const tx2 = await keypers.append();
    await tx2.wait();
    configSetIndex = lastSetIndex + 1;
  }

  const cfg = await ethers.getContract("KeyperConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlockNumber = currentBlock + 15;

  const activeConfig = await cfg.getActiveConfig(activationBlockNumber);
  if (activeConfig[1].toNumber() === configSetIndex) {
    console.log("Keyper config already added");
    return;
  }

  const tx = await cfg.addNewCfg({
    activationBlockNumber: activationBlockNumber,
    setIndex: configSetIndex,
    threshold: Math.ceil((keyperAddrs.length / 3) * 2),
  });
  await tx.wait();
  console.log(
    "configure keypers: activationBlockNumber %s, setIndex: %d, keypers: %s",
    activationBlockNumber,
    configSetIndex,
    keyperAddrs
  );
}
