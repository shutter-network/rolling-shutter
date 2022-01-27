module.exports = {
  configure_keypers: configure_keypers,
};

const hre = require("hardhat");
const ethers = hre.ethers;

// const { inspect } = require("util");

async function configure_keypers(keyperAddrs) {
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
    const addKeyperTx = await keypers.add(keyperAddrs);
    await addKeyperTx.wait(hre.numConfirmations);
    const appendTx = await keypers.append();
    await appendTx.wait(hre.numConfirmations);
    configSetIndex = lastSetIndex + 1;
  }

  const cfg = await ethers.getContract("KeyperConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlockNumber = currentBlock + 10;

  const activeConfig = await cfg.getActiveConfig(activationBlockNumber);
  if (activeConfig[1].toNumber() === configSetIndex) {
    console.log("Keyper config already added");
    return;
  }

  const addCfgTx = await cfg.addNewCfg({
    activationBlockNumber: activationBlockNumber,
    setIndex: configSetIndex,
    threshold: Math.ceil((keyperAddrs.length / 3) * 2),
  });
  await addCfgTx.wait(hre.numConfirmations);

  console.log(
    "configure keypers: activationBlockNumber %s, setIndex: %d, keypers: %s",
    activationBlockNumber,
    configSetIndex,
    keyperAddrs
  );
}
