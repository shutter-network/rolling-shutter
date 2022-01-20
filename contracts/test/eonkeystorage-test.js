const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("Empty EonKeyStorage", function () {
  let storage;
  let storageFactory;

  beforeEach(async () => {
    storageFactory = await ethers.getContractFactory("EonKeyStorage");
    storage = await storageFactory.deploy();
  });

  it("number of stored keys should be zero", async function () {
    expect(await storage.num()).to.equal(0);
  });

  it("should not return keys for any block number", async function () {
    for (const n of [0, 100]) {
      await expect(storage.get(n)).to.be.reverted;
    }
  });

  it("should successfully add first key", async function () {
    await expect(storage.insert("0x00", 100))
      .to.emit(storage, "Inserted")
      .withArgs(100, 0, "0x00");
    expect(await storage.num()).to.equal(1);
    await expect(storage.get(99)).to.be.reverted;
    expect(await storage.get(100)).to.equal("0x00");
  });

  it("should successfully add two keys in order", async function () {
    await expect(storage.insert("0x00", 100))
      .to.emit(storage, "Inserted")
      .withArgs(100, 0, "0x00");
    await expect(storage.insert("0x11", 200))
      .to.emit(storage, "Inserted")
      .withArgs(200, 1, "0x11");
    expect(await storage.num()).to.equal(2);
    await expect(storage.get(99)).to.be.reverted;
    expect(await storage.get(100)).to.equal("0x00");
    expect(await storage.get(199)).to.equal("0x00");
    expect(await storage.get(200)).to.equal("0x11");
    expect(await storage.get(10000)).to.equal("0x11");
  });

  it("should successfully add two keys out of order", async function () {
    await expect(storage.insert("0x11", 200))
      .to.emit(storage, "Inserted")
      .withArgs(200, 0, "0x11");
    await expect(storage.insert("0x00", 100))
      .to.emit(storage, "Inserted")
      .withArgs(100, 1, "0x00");
    expect(await storage.num()).to.equal(2);
    await expect(storage.get(99)).to.be.reverted;
    expect(await storage.get(100)).to.equal("0x00");
    expect(await storage.get(199)).to.equal("0x00");
    expect(await storage.get(200)).to.equal("0x11");
    expect(await storage.get(10000)).to.equal("0x11");
  });

  it("should successfully add three keys out of order", async function () {
    await expect(storage.insert("0x00", 100))
      .to.emit(storage, "Inserted")
      .withArgs(100, 0, "0x00");
    await expect(storage.insert("0x22", 300))
      .to.emit(storage, "Inserted")
      .withArgs(300, 1, "0x22");
    await expect(storage.insert("0x11", 200))
      .to.emit(storage, "Inserted")
      .withArgs(200, 2, "0x11");
    expect(await storage.num()).to.equal(3);
    await expect(storage.get(99)).to.be.reverted;
    expect(await storage.get(100)).to.equal("0x00");
    expect(await storage.get(199)).to.equal("0x00");
    expect(await storage.get(200)).to.equal("0x11");
    expect(await storage.get(299)).to.equal("0x11");
    expect(await storage.get(300)).to.equal("0x22");
    expect(await storage.get(10000)).to.equal("0x22");
  });

  it("should successfully replace keys", async function () {
    await storage.insert("0x00", 100);
    await storage.insert("0x11", 200);
    await expect(storage.insert("0x22", 200))
      .to.emit(storage, "Inserted")
      .withArgs(200, 2, "0x22");
    expect(await storage.get(199)).to.equal("0x00");
    expect(await storage.get(200)).to.equal("0x22");
    expect(await storage.get(299)).to.equal("0x22");
  });
});
