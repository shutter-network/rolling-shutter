const { expect } = require("chai");

async function deploy() {
  const Addrs = await ethers.getContractFactory("AddrsSeq");
  const aset = await Addrs.deploy();
  await aset.deployed();
  return aset;
}

describe("AddrsSeq", function () {
  it("append should increase count", async function () {
    const aset = await deploy();

    expect(await aset.count()).to.equal(0);
    await expect(aset.append()).to.emit(aset, "Appended").withArgs(0);
    expect(await aset.count()).to.equal(1);
    expect(await aset.countNth(0)).to.equal(0);

    await expect(aset.append()).to.emit(aset, "Appended").withArgs(1);
    expect(await aset.count()).to.equal(2);
  });

  it("accessing the current list should be impossible", async function () {
    const aset = await deploy();

    const a = ethers.utils.getAddress(
      "0x8ba1f109551bd432803012645ac136ddd64dba72"
    );
    tx = await aset.add([a]);
    await tx.wait();
    await expect(aset.at(0, 0)).to.be.revertedWith(
      "AddrsSeq.at: n out of range"
    );
  });

  it("append should emit an event and append the current list", async function () {
    const aset = await deploy();

    const a = ethers.utils.getAddress(
      "0x8ba1f109551bd432803012645ac136ddd64dba72"
    );
    tx = await aset.add([a]);
    await tx.wait();

    await expect(aset.append()).to.emit(aset, "Appended").withArgs(0);
    expect(await aset.count()).to.equal(1);
    expect(await aset.at(0, 0)).to.equal(a);
    expect(await aset.countNth(0)).to.equal(1);
  });
});
