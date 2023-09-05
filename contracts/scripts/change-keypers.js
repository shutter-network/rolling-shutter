#!/usr/bin/env node
/**
   This script configures a new set of keypers
 */
const fs = require("fs");
const process = require("process");
const path = require("path");
const { ethers } = require("hardhat");
const { configure_keypers } = require("../lib/configure-keypers.js");
const { fund } = require("../lib/fund.js");

// const { inspect } = require("util");

/* global __dirname */
process.chdir(path.dirname(__dirname)); // allow calling the script from anywhere
const hre = require("hardhat");

async function main() {
  if (process.env.DEPLOY_CONF === undefined) {
    console.error("please set DEPLOY_CONF environment variable");
    return;
  }
  // TODO can we get access to the hre in a script like this?
  // TODO use a different json file to determine the keyper changes
  const deployConf = JSON.parse(fs.readFileSync(process.env.DEPLOY_CONF));

  const newKeyperSet = deployConf.keypers?.at(1);
  if (newKeyperSet === undefined) {
    console.error("no updated keyper set defined in DEPLOY_CONF");
    return;
  }

  const { bank } = await hre.getNamedAccounts();
  const bankSigner = await ethers.getSigner(bank);

  const fundValue = hre.deployConf.fundValue;
  await fund(newKeyperSet, bankSigner, fundValue);
  await configure_keypers(newKeyperSet, 30);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
