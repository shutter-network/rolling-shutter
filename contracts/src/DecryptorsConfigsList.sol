// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import "./Registry.sol";
import "./AddrsSeq.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

struct DecryptorsConfig {
    uint64 activationBlockNumber;
    uint64 setIndex;
}

contract DecryptorsConfigsList is Ownable {
    DecryptorsConfig[] public decryptorsConfigs;
    AddrsSeq public addrsSeq;
    Registry public blsRegistry;

    event NewConfig(uint64 activationBlockNumber, uint64 index);

    constructor(AddrsSeq _addrsSeq, Registry _blsRegistry) {
        require(
            _addrsSeq.countNth(0) == 0,
            "AddrsSeq must have empty list at index 0"
        );
        require(
            _blsRegistry.addrsSeq() == _addrsSeq,
            "AddrsSeq of _blsRegistry must be _addrsSeq"
        );

        addrsSeq = _addrsSeq;
        blsRegistry = _blsRegistry;

        decryptorsConfigs.push(
            DecryptorsConfig({activationBlockNumber: 0, setIndex: 0})
        );
        emit NewConfig(0, 0);
    }

    function addNewCfg(DecryptorsConfig calldata config) public onlyOwner {
        require(
            addrsSeq.count() > config.setIndex,
            "No appended set in seq corresponding to config's set index"
        );
        require(
            decryptorsConfigs[decryptorsConfigs.length - 1]
                .activationBlockNumber <= config.activationBlockNumber,
            "Cannot add new set with lower block number than previous"
        );
        require(
            block.number <= config.activationBlockNumber,
            "Cannot add new set with past block number"
        );

        decryptorsConfigs.push(config);
        emit NewConfig(config.activationBlockNumber, config.setIndex);
    }

    function getActiveConfig(uint64 activationBlockNumber)
        public
        view
        returns (DecryptorsConfig memory)
    {
        for (uint256 i = decryptorsConfigs.length - 1; true; i--) {
            if (
                decryptorsConfigs[i].activationBlockNumber <=
                activationBlockNumber
            ) {
                return decryptorsConfigs[i];
            }
        }
        revert("unreachable");
    }

    function getCurrentActiveConfig()
        public
        view
        returns (DecryptorsConfig memory)
    {
        return getActiveConfig(uint64(block.number));
    }
}
