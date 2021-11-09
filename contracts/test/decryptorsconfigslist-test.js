const { expect } = require("chai");
const { ethers } = require("hardhat");

async function deployAddrsSeq() {
  const addrsSeqFactory = await ethers.getContractFactory("AddrsSeq");
  const addrsSeqContract = await addrsSeqFactory.deploy();
  await addrsSeqContract.append();
  return addrsSeqContract;
}

async function deployRegistry(addrsSeq) {
  const registryFactory = await ethers.getContractFactory("Registry");
  return registryFactory.deploy(addrsSeq);
}

describe("DecryptorsConfigsList", function () {
  it("should not be possible to deploy with inconsistent registries", async function () {
    const configsFactory = await ethers.getContractFactory(
      "DecryptorsConfigsList"
    );
    const addrsSeq1 = (await deployAddrsSeq()).address;
    const addrsSeq2 = (await deployAddrsSeq()).address;

    const registry = (await deployRegistry(addrsSeq1)).address;
    const registryWrongAddrsSeq = (await deployRegistry(addrsSeq2)).address;

    await expect(
      configsFactory.deploy(addrsSeq1, registry, registryWrongAddrsSeq)
    ).to.be.revertedWith(
      "AddrsSeq of _KeySignaturesRegistry must be _addrsSeq"
    );
    await expect(
      configsFactory.deploy(addrsSeq1, registryWrongAddrsSeq, registry)
    ).to.be.revertedWith("AddrsSeq of _BLSKeysRegistry must be _addrsSeq");
    await expect(
      configsFactory.deploy(
        addrsSeq1,
        registryWrongAddrsSeq,
        registryWrongAddrsSeq
      )
    ).to.be.reverted;
    await expect(
      configsFactory.deploy(addrsSeq1, registry, registry)
    ).to.be.revertedWith("The two used registries must be different");
  });
});
