const { expect } = require("chai");

async function deploy() {
  const addrsSeqFactory = await ethers.getContractFactory("AddrsSeq");
  const addrsSeqContract = await addrsSeqFactory.deploy();
  const keypersConfigsFactory = await ethers.getContractFactory(
    "KeypersConfigs"
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
  configContract = await configContract.getActiveConfig(blockNumber);
  return {
    activationBlockNumber: configContract[0].toNumber(),
    setIndex: configContract[1].toNumber(),
  };
}

describe("KeypersConfigs", function () {
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
    let kprSet = await configContract.keypersConfigs(1);
    expect(kprSet.activationBlockNumber).to.equal(blockNumber);
    expect(kprSet.setIndex).to.equal(index);

    await addrsSeq.append();
    expect(
      configContract.addNewCfg({
        activationBlockNumber: blockNumber2,
        setIndex: index2,
      })
    )
      .to.emit(configContract, "NewConfig")
      .withArgs(blockNumber2, index2);
    kprSet = await configContract.keypersConfigs(2);
    expect(kprSet.activationBlockNumber).to.equal(blockNumber2);
    expect(kprSet.setIndex).to.equal(index2);
  });

  it("should be impossible to add new set when not sequenced", async function () {
    const cfg = await deploy();

    await expect(cfg.addNewCfg({ activationBlockNumber: 123, setIndex: 2 })).to
      .be.reverted;
  });

  it("should be impossible to add sets in decreasing block number order", async function () {
    const cfg = await deploy();
    const blockNumber = 123;

    await cfg.addNewCfg({ activationBlockNumber: blockNumber, setIndex: 0 });
    await expect(
      cfg.addNewCfg({ activationBlockNumber: blockNumber - 1, setIndex: 0 })
    ).to.be.reverted;
  });

  it("not owner should not be able to add new set", async function () {
    const cfg = await deploy();

    const notOwner = (await hre.ethers.getSigners())[1];
    expect(cfg.owner()).to.not.be.equal(notOwner);

    await expect(
      cfg
        .connect(notOwner)
        .addNewCfg({ activationBlockNumber: 123, setIndex: 0 })
    ).to.be.reverted;
  });

  it("should return active config", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);

    await expect((await getConfig(cfg, 0)).activationBlockNumber).to.equal(0);
    await expect((await getConfig(cfg, 0)).setIndex).to.equal(0);

    const blockNumber = 123;
    const index = 1;
    await addrsSeq.append();
    await cfg.addNewCfg({
      activationBlockNumber: blockNumber,
      setIndex: index,
    });

    await expect(
      (
        await getConfig(cfg, blockNumber)
      ).activationBlockNumber
    ).to.equal(blockNumber);
    await expect((await getConfig(cfg, blockNumber)).setIndex).to.equal(index);

    await expect(
      (
        await getConfig(cfg, blockNumber + 1)
      ).activationBlockNumber
    ).to.equal(blockNumber);
    await expect((await getConfig(cfg, blockNumber + 1)).setIndex).to.equal(
      index
    );

    await expect(
      (
        await getConfig(cfg, blockNumber - 1)
      ).activationBlockNumber
    ).to.equal(0);
    await expect((await getConfig(cfg, blockNumber - 1)).setIndex).to.equal(0);
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

    await expect((await getConfig(cfg, blockNumber)).setIndex).to.equal(
      newIndex
    );
  });
});
