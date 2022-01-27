const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var collatorAddress = [await hre.getCollatorAddress()];

  const collator = await ethers.getContract("Collator");
  const lastSetIndex = (await collator.count()) - 1;
  let configSetIndex;

  let setAdded = true;
  if (
    (await collator.countNth(lastSetIndex)).toNumber() !==
    collatorAddress.length
  ) {
    setAdded = false;
  } else {
    for (const i of Array(collatorAddress.length).keys()) {
      if ((await collator.at(lastSetIndex, i)) !== collatorAddress[i]) {
        setAdded = false;
        break;
      }
    }
  }
  if (setAdded) {
    console.log("Collator set already added");
    configSetIndex = lastSetIndex;
  } else {
    const addCollatorTx = await collator.add(collatorAddress);
    await addCollatorTx.wait(hre.numConfirmations);
    const appendTx = await collator.append();
    await appendTx.wait(hre.numConfirmations);
    configSetIndex = lastSetIndex + 1;
  }

  const cfg = await ethers.getContract("CollatorConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlockNumber = currentBlock + 10;

  const activeConfig = await cfg.getActiveConfig(activationBlockNumber);
  if (activeConfig[1].toNumber() === configSetIndex) {
    console.log("Collator config already added");
    return;
  }

  const addCfgTx = await cfg.addNewCfg({
    activationBlockNumber: activationBlockNumber,
    setIndex: configSetIndex,
  });
  await addCfgTx.wait(hre.numConfirmations);

  console.log(
    "configure collator: activationBlockNumber %s collator: %s",
    activationBlockNumber,
    ...collatorAddress
  );
};
