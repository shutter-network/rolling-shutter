require("@nomiclabs/hardhat-waffle");
require("@eth-optimism/hardhat-ovm");
require("hardhat-deploy");
require("hardhat-deploy-ethers");
require("@nomiclabs/hardhat-ganache");
const { task, extendEnvironment } = require("hardhat/config");
const fs = require("fs");
const process = require("process");

// This is a sample Hardhat task. To learn how to create your own go to
// https://hardhat.org/guides/create-task.html
task("accounts", "Prints the list of accounts", async (taskArgs, hre) => {
  const accounts = await hre.ethers.getSigners();

  for (const account of accounts) {
    console.log(account.address);
  }
});

extendEnvironment((hre) => {
  if (process.env.DEPLOY_CONF !== undefined) {
    hre.deployConf = JSON.parse(fs.readFileSync(process.env.DEPLOY_CONF));
  } else {
    hre.deployConf = {
      keypers: null,
      decryptors: null,
      collator: null,
      fundValue: "",
    };
  }

  hre.getKeyperAddresses = async function () {
    if (hre.deployConf.keypers === null) {
      const { keyper0, keyper1, keyper2 } = await hre.getNamedAccounts();
      return [keyper0, keyper1, keyper2];
    } else {
      return hre.deployConf.keypers;
    }
  };

  hre.getDecryptorAddresses = async function () {
    if (hre.deployConf.decryptors === null) {
      const { decryptor0, decryptor1, decryptor2 } =
        await hre.getNamedAccounts();
      return [decryptor0, decryptor1, decryptor2];
    } else {
      return hre.deployConf.decryptors;
    }
  };

  hre.getCollatorAddress = async function () {
    if (hre.deployConf.collator === null) {
      const { collator } = await hre.getNamedAccounts();
      return collator;
    } else {
      return hre.deployConf.collator;
    }
  };

  // the number of confirmations to wait when sending transactions
  if (process.env.NUM_CONFIRMATIONS) {
    hre.numConfirmations = parseInt(process.env.NUM_CONFIRMATIONS);
  } else {
    hre.numConfirmations = 0;
  }
});

// You need to export an object to set up your config
// Go to https://hardhat.org/config/ to learn more

/**
 * @type import('hardhat/config').HardhatUserConfig
 */
module.exports = {
  solidity: "0.8.9",
  paths: {
    sources: "./src",
  },
  namedAccounts: {
    deployer: 0,
    keyper0: 1,
    keyper1: 2,
    keyper2: 3,
    decryptor0: 4,
    decryptor1: 5,
    decryptor2: 6,
    collator: 7,
  },
  networks: {
    hardhat: {
      mining: {
        auto: true,
        interval: 1500,
      },
    },
    optimistic: {
      url: "http://127.0.0.1:8545",
      accounts: {
        mnemonic: "test test test test test test test test test test test junk",
      },
      gasPrice: 15000000,
      ovm: true, // This sets the network as using the ovm and ensure contract will be compiled against that.
    },
  },
};
