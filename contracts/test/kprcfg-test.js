const { expect } = require("chai");

async function deploy() {
  const addrsSeqF = await ethers.getContractFactory("AddrsSeq");
  const addrsSeq = await addrsSeqF.deploy();
  const cfgF = await ethers.getContractFactory("KprCfgs");
  const cfg = await cfgF.deploy(addrsSeq.address);
  await cfg.deployed();
  return cfg;
}

async function getAddrsSeq(cfg) {
  const addrs = await cfg.addrsSeq();
  const addrsSeqF = await ethers.getContractFactory("AddrsSeq");
  return addrsSeqF.attach(addrs);
}

async function getCfg(cfg, b) {
  cfg = await cfg.getActiveCfg(b);
  return { blkNbr: cfg[0].toNumber(), index: cfg[1].toNumber() };
}

const a = ethers.utils.getAddress("0x8ba1f109551bd432803012645ac136ddd64dba72");

describe("KprCfg", function () {
  it("adding new set should emit an event", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);

    const blkNmbr = 123;
    const index = 0;
    const blkNmbr2 = 123456;
    const index2 = 1;

    await expect(cfg.addNewCfg({ blkNbr: blkNmbr, index: index }))
      .to.emit(cfg, "NewCfg")
      .withArgs(blkNmbr, index);
    let kprSet = await cfg.kprCfgs(1);
    expect(kprSet.blkNbr).to.equal(blkNmbr);
    expect(kprSet.index).to.equal(index);

    await addrsSeq.append();
    expect(cfg.addNewCfg({ blkNbr: blkNmbr2, index: index2 }))
      .to.emit(cfg, "NewCfg")
      .withArgs(blkNmbr2, index2);
    kprSet = await cfg.kprCfgs(2);
    expect(kprSet.blkNbr).to.equal(blkNmbr2);
    expect(kprSet.index).to.equal(index2);
  });

  it("should be impossible to add new set when not sequenced", async function () {
    const cfg = await deploy();

    await expect(cfg.addNewCfg({ blkNbr: 123, index: 2 })).to.be.reverted;
  });

  it("should be impossible to add sets in decreasing block number order", async function () {
    const cfg = await deploy();
    const blkNbr = 123;

    await cfg.addNewCfg({ blkNbr: blkNbr, index: 0 });
    await expect(cfg.addNewCfg({ blkNbr: blkNbr - 1, index: 0 })).to.be
      .reverted;
  });

  it("not owner should not be able to add new set", async function () {
    const cfg = await deploy();

    await expect(cfg.connect(a).addNewCfg({ blkNbr: 123, index: 0 })).to.be
      .reverted;
  });

  it("should return active config", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);

    await expect((await getCfg(cfg, 0)).blkNbr).to.equal(0);
    await expect((await getCfg(cfg, 0)).index).to.equal(0);

    const blkNmbr = 123;
    const index = 1;
    await addrsSeq.append();
    await cfg.addNewCfg({ blkNbr: blkNmbr, index: index });

    await expect((await getCfg(cfg, blkNmbr)).blkNbr).to.equal(blkNmbr);
    await expect((await getCfg(cfg, blkNmbr)).index).to.equal(index);

    await expect((await getCfg(cfg, blkNmbr + 1)).blkNbr).to.equal(blkNmbr);
    await expect((await getCfg(cfg, blkNmbr + 1)).index).to.equal(index);

    await expect((await getCfg(cfg, blkNmbr - 1)).blkNbr).to.equal(0);
    await expect((await getCfg(cfg, blkNmbr - 1)).index).to.equal(0);
  });

  it("should replace current set with same block number", async function () {
    const cfg = await deploy();
    const addrsSeq = await getAddrsSeq(cfg);
    const blkNbr = 123;
    const newIndex = 1;

    await cfg.addNewCfg({ blkNbr: blkNbr, index: 0 });
    await addrsSeq.append();
    await cfg.addNewCfg({ blkNbr: blkNbr, index: newIndex });

    await expect((await getCfg(cfg, blkNbr)).index).to.equal(newIndex);
  });
});
