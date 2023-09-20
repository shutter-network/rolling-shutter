require("@nomiclabs/hardhat-waffle");
require("@eth-optimism/hardhat-ovm");
require("hardhat-deploy");
require("hardhat-deploy-ethers");
require("@nomiclabs/hardhat-ganache");
require("@nomicfoundation/hardhat-verify");

const { task, extendEnvironment } = require("hardhat/config");
const fs = require("fs");
const process = require("process");

let etherscanAPIKey = {};
if (process.env.GNOSISSCAN_API_KEY !== undefined) {
  etherscanAPIKey.gnosis = process.env.GNOSISSCAN_API_KEY;
  // console.log("Using etherscan API key configured via GNOSISSCAN_API_KEY");
}

let gnosisAccounts = [];
if (process.env.GNOSIS_DEPLOY_KEY !== undefined) {
  gnosisAccounts.push(process.env.GNOSIS_DEPLOY_KEY);
  // console.log("Using gnosis deploy key configured via GNOSIS_DEPLOY_KEY");
}

// This is a sample Hardhat task. To learn how to create your own go to
// https://hardhat.org/guides/create-task.html
task("accounts", "Prints the list of accounts", async (taskArgs, hre) => {
  const accounts = await hre.ethers.getSigners();

  for (const account of accounts) {
    console.log(account.address);
  }
});

task("verify-all", "Verifies all contracts", async (taskArgs, hre) => {
  const keypers = await hre.ethers.getContract("Keypers");
  console.log("Verifying Keypers @ %s", keypers.address);
  await hre.run("verify:verify", {
    address: keypers.address,
    constructorArguments: [],
  });

  const keyperConfig = await hre.ethers.getContract("KeyperConfig");
  console.log("Verifying KeyperConfig @ %s", keyperConfig.address);
  await hre.run("verify:verify", {
    address: keyperConfig.address,
    constructorArguments: [keypers.address],
  });

  const collator = await hre.ethers.getContract("Collator");
  console.log("Verifying Collator @ %s", collator.address);
  await hre.run("verify:verify", {
    address: collator.address,
    constructorArguments: [],
  });

  const collatorConfig = await hre.ethers.getContract("CollatorConfig");
  console.log("Verifying CollatorConfig @ %s", collatorConfig.address);
  await hre.run("verify:verify", {
    address: collatorConfig.address,
    constructorArguments: [collator.address],
  });

  const eonKeyStorage = await hre.ethers.getContract("EonKeyStorage");
  console.log("Verifying EonKeyStorage @ %s", eonKeyStorage.address);
  await hre.run("verify:verify", {
    address: eonKeyStorage.address,
    constructorArguments: [],
  });

  const batchCounter = await hre.ethers.getContract("BatchCounter");
  console.log("Verifying BatchCounter @ %s", batchCounter.address);
  await hre.run("verify:verify", {
    address: batchCounter.address,
    constructorArguments: [],
  });
});

extendEnvironment((hre) => {
  if (process.env.DEPLOY_CONF !== undefined) {
    hre.deployConf = JSON.parse(fs.readFileSync(process.env.DEPLOY_CONF));
  } else {
    hre.deployConf = {
      keypers: null,
      collator: null,
      fundValue: "",
      activationBlockOffset: 30,
      thresholdRatio: 2 / 3,
    };
  }

  hre.getKeyperAddresses = async function (index = -1) {
    if (index === -1) {
      index = parseInt(process.env.KEYPER_SET_INDEX ?? 0, 10);
    }
    const keypers = hre.deployConf.keypers?.at(index);
    if (keypers === undefined) {
      const { keyper0, keyper1, keyper2 } = await hre.getNamedAccounts();
      if (keyper0 && keyper1 && keyper2) {
        return [keyper0, keyper1, keyper2];
      } else {
        return [];
      }
    } else {
      return keypers;
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
    keyper3: 4,
    collator: 7,
    bank: {
      // an account that has funds
      default: 0,
      nitro: 1,
    },
  },
  networks: {
    hardhat: {
      mining: {
        auto: true,
        interval: 1500,
      },
    },
    gnosis: {
      url: "https://rpc.gnosis.gateway.fm",
      accounts: gnosisAccounts,
    },
    nitro: {
      url: "http://localhost:8547",
      accounts: [
        "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", // first hardhat acccount
        "0xe887f7d17d07cc7b8004053fb8826f6657084e88904bb61590e498ca04704cf2", // nitro funnel
      ],
    },
    dockerGeth: {
      url: "http://geth:8545",
      accounts: "remote",
    },
  },
  etherscan: {
    customChains: [
      {
        network: "gnosis",
        chainId: 100,
        urls: {
          apiURL: "https://api.gnosisscan.io/api",
          browserURL: "https://gnosisscan.io/",
        },
      },
    ],
    apiKey: etherscanAPIKey,
  },
};
