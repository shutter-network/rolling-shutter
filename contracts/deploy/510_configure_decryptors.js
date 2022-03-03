const { ethers } = require("hardhat");

module.exports = async function (hre) {
  var decryptorAddrs = await hre.getDecryptorAddresses();

  const decryptors = await ethers.getContract("Decryptors");
  const lastSetIndex = (await decryptors.count()) - 1;
  let configSetIndex;

  let setAdded = true;
  if (
    (await decryptors.countNth(lastSetIndex)).toNumber() !==
    decryptorAddrs.length
  ) {
    setAdded = false;
  } else {
    for (const i of Array(decryptorAddrs.length).keys()) {
      if ((await decryptors.at(lastSetIndex, i)) !== decryptorAddrs[i]) {
        setAdded = false;
        break;
      }
    }
  }
  if (setAdded) {
    console.log("Decryptor set already added");
    configSetIndex = lastSetIndex;
  } else {
    await decryptors.add(decryptorAddrs);
    await decryptors.append();
    configSetIndex = lastSetIndex + 1;
  }

  const cfg = await ethers.getContract("DecryptorConfig");
  const currentBlock = await ethers.provider.getBlockNumber();
  const activationBlockNumber = currentBlock + 10;

  const activeConfig = await cfg.getActiveConfig(activationBlockNumber);
  if (activeConfig[1].toNumber() === configSetIndex) {
    console.log("Decryptor config already added");
    return;
  }

  await cfg.addNewCfg({
    activationBlockNumber: activationBlockNumber,
    setIndex: configSetIndex,
  });
  console.log(
    "configure decryptors: activationBlockNumber %s decryptors: %s",
    activationBlockNumber,
    decryptorAddrs
  );
};
