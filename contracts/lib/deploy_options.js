const hre = require("hardhat");

module.exports = {
  log: true,
  waitConfirmations: hre.numConfirmations,
};
