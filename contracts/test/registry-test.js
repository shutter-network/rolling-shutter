const { expect } = require("chai");
const { ethers } = require("hardhat");

async function deployAddrsSeq() {
  const Addrs = await ethers.getContractFactory("AddrsSeq");
  const aset = await Addrs.deploy();
  await aset.deployed();
  return aset;
}

async function deploy(aset) {
  const Addrs = await ethers.getContractFactory("Registry");
  const r = await Addrs.deploy(aset.address);
  await r.deployed();
  return r;
}

describe("Registry", function () {
  let aset;
  let registry;
  let owner;
  const a = ethers.utils.getAddress(
    "0x8ba1f109551bd432803012645ac136ddd64dba72"
  );
  const v = new Uint8Array([1, 2, 3, 4]);

  beforeEach(async function () {
    [owner] = await ethers.getSigners();
    aset = await deployAddrsSeq();
    registry = await deploy(aset);
  });

  it("non-members should not be able to register", async function () {
    var tx = await aset.add([a]);
    await tx.wait();
    tx = await aset.append();
    await tx.wait();
    await expect(registry.register(0, 0, v)).to.be.revertedWith(
      "Registry: sender is not allowed"
    );
  });

  it("members can call register exactly once", async function () {
    var tx = await aset.add([owner.address]);
    await tx.wait();
    tx = await aset.append();
    await tx.wait();
    expect(await registry.get(owner.address)).to.equal("0x");

    tx = await registry.register(0, 0, v);
    await tx.wait();
    expect(await registry.get(owner.address)).to.equal("0x01020304");

    await expect(registry.register(0, 0, v)).to.be.revertedWith(
      "Registry: sender already registered"
    );
  });

  it("members cannot register empty value", async function () {
    var tx = await aset.add([owner.address]);
    await tx.wait();
    tx = await aset.append();
    await tx.wait();
    expect(await registry.get(owner.address)).to.equal("0x");

    await expect(
      registry.register(0, 0, new Uint8Array([]))
    ).to.be.revertedWith("Registry: cannot register empty value");
  });
});
