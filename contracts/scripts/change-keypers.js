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

  if (process.env.KEYPER_SET_INDEX === undefined) {
    console.error("please set KEYPER_SET_INDEX environment variable");
    return;
  }

  const keyperSetIndex = parseInt(process.env.KEYPER_SET_INDEX);
  const deployConf = JSON.parse(fs.readFileSync(process.env.DEPLOY_CONF));

  const newKeyperSet = deployConf.keypers?.at(keyperSetIndex);
  if (newKeyperSet === undefined) {
    console.error(
      "no updated keyper set defined in DEPLOY_CONF at index %s",
      keyperSetIndex
    );
    return;
  }

  const { bank } = await hre.getNamedAccounts();
  const bankSigner = await ethers.getSigner(bank);

  const fundValue = hre.deployConf.fundValue;
  const activationBlockOffset = hre.deployConf.activationBlockOffset;
  await fund(newKeyperSet, bankSigner, fundValue);
  await configure_keypers(newKeyperSet, activationBlockOffset);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
