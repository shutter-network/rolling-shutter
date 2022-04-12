const { expect } = require("chai");
const { ethers } = require("hardhat");

const addresses = [
  "0x0000000000000000000000000000000000000000",
  "0x1111111111111111111111111111111111111111",
  "0x2222222222222222222222222222222222222222",
  "0x3333333333333333333333333333333333333333",
];

async function deploy() {
  const addrsSeqFactory = await ethers.getContractFactory("AddrsSeq");
  const addrsSeqContract = await addrsSeqFactory.deploy();
  await addrsSeqContract.append();
  const keypersConfigsFactory = await ethers.getContractFactory(
    "KeypersConfigsList"
  );
  const keypersConfigsContract = await keypersConfigsFactory.deploy(
    addrsSeqContract.address
  );
  await keypersConfigsContract.deployed();
  return keypersConfigsContract;
}

async function getAddrsSeq(configContract) {
  const address = await configContract.addrsSeq();
  const addrsSeqFactory = await ethers.getContractFactory("AddrsSeq");
  return addrsSeqFactory.attach(address);
}

async function getConfig(configContract, blockNumber) {
  const config = await configContract.getActiveConfig(blockNumber);
  return {
    activationBlockNumber: config.activationBlockNumber.toNumber(),
    setIndex: config.setIndex.toNumber(),
    threshold: config.threshold.toNumber(),
  };
}

describe("KeypersConfigsList", function () {
  it("adding new set should emit an event", async function () {
    const configContract = await deploy();
    const addrsSeq = await getAddrsSeq(configContract);

    const blockNumber = 123;
    const setIndex = 0;
    const threshold = 0;
    const blockNumber2 = 123456;
    const setIndex2 = 1;
    const threshold2 = 3;

    await expect(
      configContract.addNewCfg({
        activationBlockNumber: blockNumber,
        setIndex: setIndex,
        threshold: threshold,
      })
    )
      .to.emit(configContract, "NewConfig")
      .withArgs(blockNumber, setIndex, 1, threshold);
    let kprSet = await configContract.keypersConfigs(1);
    expect(kprSet.activationBlockNumber).to.equal(blockNumber);
    expect(kprSet.setIndex).to.equal(setIndex);
    expect(kprSet.threshold).to.equal(threshold);

    await addrsSeq.add(addresses);
    await addrsSeq.append();
    await expect(
      configContract.addNewCfg({
        activationBlockNumber: blockNumber2,
        setIndex: setIndex2,
        threshold: threshold2,
      })
    )
      .to.emit(configContract, "NewConfig")
      .withArgs(blockNumber2, setIndex2, 2, threshold2);
    kprSet = await configContract.keypersConfigs(2);
    expect(kprSet.activationBlockNumber).to.equal(blockNumber2);
    expect(kprSet.setIndex).to.equal(setIndex2);
    expect(kprSet.threshold).to.equal(threshold2);
  });

  it("should be impossible to add new set when not sequenced", async function () {
    const cfg = await deploy();

    await expect(
      cfg.addNewCfg({ activationBlockNumber: 123, setIndex: 2, threshold: 1 })
    ).to.be.revertedWith(
      "No appended set in seq corresponding to config's set index"
    );
  });

  it("should be impossible to add sets in decreasing block number order", async function () {
    const cfg = await deploy();
    const blockNumber = 123;

    await cfg.addNewCfg({
      activationBlockNumber: blockNumber,
      setIndex: 0,
      threshold: 0,
    });
    await expect(
      cfg.addNewCfg({
        activationBlockNumber: blockNumber - 1,
        setIndex: 0,
        threshold: 0,
      })
    ).to.be.revertedWith(
      "Cannot add new set with lower block number than previous"
    );
  });

  it("should be impossible to add sets with past block number", async function () {
    const cfg = await deploy();

    await expect(
      cfg.addNewCfg({ activationBlockNumber: 1, setIndex: 0, threshold: 0 })
    ).to.be.revertedWith("Cannot add new set with past block number");
  });

  it("should be impossible to add sets with threshold 0 and non-empty keyper set", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);

    await addrsSeq.add(addresses);
    await addrsSeq.append();

    await expect(
      cfg.addNewCfg({ activationBlockNumber: 123, setIndex: 1, threshold: 0 })
    ).to.be.revertedWith("Threshold must be at least one");
  });

  it("should be impossible to add sets with threshold non-zero and empty keyper set", async function () {
    const cfg = await deploy();

    await expect(
      cfg.addNewCfg({ activationBlockNumber: 123, setIndex: 0, threshold: 1 })
    ).to.be.revertedWith("Threshold must be zero if keyper set is empty");
  });

  it("should be impossible to add sets with threshold exceeding keyper set size", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);

    await addrsSeq.add(addresses);
    await addrsSeq.append();

    await expect(
      cfg.addNewCfg({ activationBlockNumber: 123, setIndex: 1, threshold: 5 })
    ).to.be.revertedWith("Threshold must not exceed keyper set size");
  });

  it("not owner should not be able to add new set", async function () {
    const cfg = await deploy();

    const notOwner = (await ethers.getSigners())[1];
    expect(cfg.owner()).to.not.be.equal(notOwner);

    await expect(
      cfg
        .connect(notOwner)
        .addNewCfg({ activationBlockNumber: 123, setIndex: 0, threshold: 1 })
    ).to.be.revertedWith("Ownable: caller is not the owner");
  });

  it("should return active config", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);

    expect(await getConfig(cfg, 0)).to.deep.equal({
      activationBlockNumber: 0,
      setIndex: 0,
      threshold: 0,
    });

    const blockNumber = 123;
    const setIndex = 1;
    const threshold = 2;
    await addrsSeq.add(addresses);
    await addrsSeq.append();
    await cfg.addNewCfg({
      activationBlockNumber: blockNumber,
      setIndex: setIndex,
      threshold: threshold,
    });

    expect(await getConfig(cfg, blockNumber)).to.deep.equal({
      activationBlockNumber: blockNumber,
      setIndex: setIndex,
      threshold: threshold,
    });
    expect(await getConfig(cfg, blockNumber + 1)).to.deep.equal({
      activationBlockNumber: blockNumber,
      setIndex: setIndex,
      threshold: threshold,
    });
    expect(await getConfig(cfg, blockNumber - 1)).to.deep.equal({
      activationBlockNumber: 0,
      setIndex: 0,
      threshold: 0,
    });
  });

  it("should replace current set with same block number", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);
    const blockNumber = 123;
    const newIndex = 1;
    const threshold = 1;

    await cfg.addNewCfg({
      activationBlockNumber: blockNumber,
      setIndex: 0,
      threshold: 0,
    });
    await addrsSeq.add(addresses);
    await addrsSeq.append();
    await cfg.addNewCfg({
      activationBlockNumber: blockNumber,
      setIndex: newIndex,
      threshold: threshold,
    });

    expect(await getConfig(cfg, blockNumber)).to.deep.equal({
      activationBlockNumber: blockNumber,
      setIndex: newIndex,
      threshold: threshold,
    });
  });
});
