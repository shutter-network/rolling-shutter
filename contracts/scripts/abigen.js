#! /usr/bin/env nodejs
/**
   This script compiles the contracts for subsequent use with go-ethereum's abigen tool.
   It creates a combined.json file in the current directory.
 */

const hre = require("hardhat");
const { spawnSync } = require("child_process");
const { inspect } = require("util");

async function main() {
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
    console.log(child.stdout.toString());
    console.log(child.stderr.toString());
    throw Error(child.error);
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
