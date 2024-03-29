const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var collatorAddress = [await hre.getCollatorAddress()];
  if (!collatorAddress[0]) {
    console.log("WARNING: cannot confgure collator: address not set");
    return;
  }
  const collator = await ethers.getContract("Collator");
  const lastSetIndex = (await collator.count()) - 1;
  let configSetIndex;

  let setAdded = true;
  if (
    (await collator.countNth(lastSetIndex)).toNumber() !==
    collatorAddress.length
  ) {
    console.log("setting setAdded = false;");
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
    const tx1 = await collator.add(collatorAddress);
    await tx1.wait();
    const tx2 = await collator.append();
    await tx2.wait();
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

  const tx = await cfg.addNewCfg({
    activationBlockNumber: activationBlockNumber,
    setIndex: configSetIndex,
  });
  await tx.wait();
  console.log(
    "configure collator: activationBlockNumber %s collator: %s",
    activationBlockNumber,
    ...collatorAddress
  );
};
