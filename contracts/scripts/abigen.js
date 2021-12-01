#!/usr/bin/env node
/**
   This script compiles the contracts and generates bindings with go-ethereum's abigen tool.
   It creates a combined.json file in the current directory.
 */
const path = require("path");
const process = require("process");
const { spawnSync } = require("child_process");
// const { inspect } = require("util");

/* global __dirname */
process.chdir(path.dirname(__dirname)); // allow calling the script from anywhere
const hre = require("hardhat");

async function run_solc() {
  const solcVersion = hre.config.solidity.compilers[0].version;
  const solc = await hre.run("compile:solidity:solc:get-build", {
    quiet: false,
    solcVersion: solcVersion,
  });
  // console.log(inspect(solc));

  if (solc.isSolcJs) {
    throw Error("Cannot use solcjs");
  }

  // console.log("Using %s", solc.compilerPath);
  const child = spawnSync(solc.compilerPath, [
    "--allow-paths=.",
    "hardhat=./node_modules/hardhat",
    "@openzeppelin=./node_modules/@openzeppelin",
    "--combined-json=bin,bin-runtime,ast,metadata,abi,srcmap,srcmap-runtime,storage-layout",
    "--optimize",
    "--overwrite",
    "--output-dir=.",
    "src/binding.sol",
  ]);

  if (child.error != undefined) {
    throw Error(child.error);
  }
  if (child.status != 0) {
    console.log(child.stderr.toString());
    throw Error("solc exited with non-zero exit status");
  }
}

function run_abigen() {
  process.chdir("../rolling-shutter");
  const child = spawnSync("abigen", [
    "--pkg",
    "contract",
    "--out",
    "contract/binding.abigen.gen.go",
    "--combined-json",
    "../contracts/combined.json",
  ]);
  if (child.error != undefined) {
    throw Error(child.error);
  }
  if (child.status != 0) {
    console.log(child.stderr.toString());
    throw Error("abigen exited with non-zero exit status");
  }
}

async function main() {
  await run_solc();
  run_abigen();
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
