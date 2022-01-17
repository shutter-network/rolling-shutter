#!/usr/bin/env node
/**
   This script configures a new set of keypers
 */
const fs = require("fs");
const process = require("process");
const path = require("path");
const { configure_keypers } = require("../lib/configure-keypers.js");

// const { inspect } = require("util");

/* global __dirname */
process.chdir(path.dirname(__dirname)); // allow calling the script from anywhere

async function main() {
  const deployConf = JSON.parse(fs.readFileSync(process.env.DEPLOY_CONF));
  await configure_keypers(deployConf.keypers);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
