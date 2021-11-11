const { expect } = require("chai");
const { ethers } = require("hardhat");

async function deploy() {
  const addrsSeqFactory = await ethers.getContractFactory("AddrsSeq");
  const addrsSeqContract = await addrsSeqFactory.deploy();
  await addrsSeqContract.append();
  const SetConfigFactory = await ethers.getContractFactory("SetConfigsList");
  const SetConfig = await SetConfigFactory.deploy(addrsSeqContract.address);
  await SetConfig.deployed();
  return SetConfig;
}

async function getAddrsSeq(configContract) {
  const address = await configContract.addrsSeq();
  const addrsSeqFactory = await ethers.getContractFactory("AddrsSeq");
  return addrsSeqFactory.attach(address);
}

async function getConfig(configContract, blockNumber) {
  configContract = await configContract.getActiveConfig(blockNumber);
  return {
    activationBlockNumber: configContract[0].toNumber(),
    setIndex: configContract[1].toNumber(),
  };
}

describe("SetConfigList", function () {
  it("adding new set should emit an event", async function () {
    const configContract = await deploy();
    const addrsSeq = await getAddrsSeq(configContract);

    const blockNumber = 123;
    const index = 0;
    const blockNumber2 = 123456;
    const index2 = 1;

    await expect(
      configContract.addNewCfg({
        activationBlockNumber: blockNumber,
        setIndex: index,
      })
    )
      .to.emit(configContract, "NewConfig")
      .withArgs(blockNumber, index);
    let setConfig = await configContract.configs(1);
    expect(setConfig.activationBlockNumber).to.equal(blockNumber);
    expect(setConfig.setIndex).to.equal(index);

    await addrsSeq.append();
    await expect(
      configContract.addNewCfg({
        activationBlockNumber: blockNumber2,
        setIndex: index2,
      })
    )
      .to.emit(configContract, "NewConfig")
      .withArgs(blockNumber2, index2);
    setConfig = await configContract.configs(2);
    expect(setConfig.activationBlockNumber).to.equal(blockNumber2);
    expect(setConfig.setIndex).to.equal(index2);
  });

  it("should be impossible to add new set when not sequenced", async function () {
    const cfg = await deploy();

    await expect(
      cfg.addNewCfg({ activationBlockNumber: 123, setIndex: 2 })
    ).to.be.revertedWith(
      "No appended set in seq corresponding to config's set index"
    );
  });

  it("should be impossible to add sets in decreasing block number order", async function () {
    const cfg = await deploy();
    const blockNumber = 123;

    await cfg.addNewCfg({ activationBlockNumber: blockNumber, setIndex: 0 });
    await expect(
      cfg.addNewCfg({ activationBlockNumber: blockNumber - 1, setIndex: 0 })
    ).to.be.revertedWith(
      "Cannot add new set with lower block number than previous"
    );
  });

  it("should be impossible to add sets with past block number", async function () {
    const cfg = await deploy();

    await expect(
      cfg.addNewCfg({ activationBlockNumber: 1, setIndex: 0 })
    ).to.be.revertedWith("Cannot add new set with past block number");
  });

  it("not owner should not be able to add new set", async function () {
    const cfg = await deploy();

    const notOwner = (await ethers.getSigners())[1];
    expect(cfg.owner()).to.not.be.equal(notOwner);

    await expect(
      cfg
        .connect(notOwner)
        .addNewCfg({ activationBlockNumber: 123, setIndex: 0 })
    ).to.be.revertedWith("Ownable: caller is not the owner");
  });

  it("should return active config", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);

    expect(await getConfig(cfg, 0)).to.deep.equal({
      activationBlockNumber: 0,
      setIndex: 0,
    });

    const blockNumber = 123;
    const index = 1;
    await addrsSeq.append();
    await cfg.addNewCfg({
      activationBlockNumber: blockNumber,
      setIndex: index,
    });

    expect(await getConfig(cfg, blockNumber)).to.deep.equal({
      activationBlockNumber: blockNumber,
      setIndex: index,
    });
    expect(await getConfig(cfg, blockNumber + 1)).to.deep.equal({
      activationBlockNumber: blockNumber,
      setIndex: index,
    });
    expect(await getConfig(cfg, blockNumber - 1)).to.deep.equal({
      activationBlockNumber: 0,
      setIndex: 0,
    });
  });

  it("should replace current set with same block number", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);
    const blockNumber = 123;
    const newIndex = 1;

    await cfg.addNewCfg({ activationBlockNumber: blockNumber, setIndex: 0 });
    await addrsSeq.append();
    await cfg.addNewCfg({
      activationBlockNumber: blockNumber,
      setIndex: newIndex,
    });

    expect(await getConfig(cfg, blockNumber)).to.deep.equal({
      activationBlockNumber: blockNumber,
      setIndex: newIndex,
    });
  });
});
