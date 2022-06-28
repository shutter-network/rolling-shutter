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
      collator: null,
      fundValue: "",
    };
  }

  hre.getKeyperAddresses = async function () {
    if (hre.deployConf.keypers === null) {
      const { keyper0, keyper1, keyper2 } = await hre.getNamedAccounts();
      if (keyper0 && keyper1 && keyper2) {
        return [keyper0, keyper1, keyper2];
      } else {
        return [];
      }
    } else {
      return hre.deployConf.keypers;
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
    nitro: {
      url: "http://127.0.0.1:8547",
      accounts: [
        "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
      ],
    },
  },
};
