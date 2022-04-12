#!/usr/bin/env node

/**
   This script adds a spare unused keyper set to the AddrSeq contract holding the keyper sets.
   see https://github.com/shutter-network/rolling-shutter/issues/240
 */
const process = require("process");
const path = require("path");
const { ethers } = require("hardhat");

// const { inspect } = require("util");

/* global __dirname */
process.chdir(path.dirname(__dirname)); // allow calling the script from anywhere

async function main() {
  const keyperAddrs = [
    "0xb7f85e85201fd635f139bc0f1d1133c1bb0a1800",
    "0xb7f85e85201fd635f139bc0f1d1133c1bb0a1801",
    "0xb7f85e85201fd635f139bc0f1d1133c1bb0a1802",
    "0xb7f85e85201fd635f139bc0f1d1133c1bb0a1803",
  ];
  const keypers = await ethers.getContract("Keypers");
  await keypers.add(keyperAddrs);
  await keypers.append();
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
