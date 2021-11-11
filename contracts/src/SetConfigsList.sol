// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import "./AddrsSeq.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

struct SetConfig {
    uint64 activationBlockNumber;
    uint64 setIndex;
}

contract SetConfigsList is Ownable {
    SetConfig[] public configs;
    AddrsSeq public addrsSeq;

    event NewConfig(uint64 activationBlockNumber, uint64 index);

    constructor(AddrsSeq _addrsSeq) {
        require(
            _addrsSeq.countNth(0) == 0,
            "AddrsSeq must have empty list at index 0"
        );

        addrsSeq = _addrsSeq;

        configs.push(SetConfig({activationBlockNumber: 0, setIndex: 0}));
        emit NewConfig(0, 0);
    }

    function addNewCfg(SetConfig calldata config) public onlyOwner {
        require(
            addrsSeq.count() > config.setIndex,
            "No appended set in seq corresponding to config's set index"
        );
        require(
            configs[configs.length - 1].activationBlockNumber <=
                config.activationBlockNumber,
            "Cannot add new set with lower block number than previous"
        );
        require(
            block.number <= config.activationBlockNumber,
            "Cannot add new set with past block number"
        );

        configs.push(config);
        emit NewConfig(config.activationBlockNumber, config.setIndex);
    }

    function getActiveConfig(uint64 activationBlockNumber)
        public
        view
        returns (SetConfig memory)
    {
        for (uint256 i = configs.length - 1; true; i--) {
            if (configs[i].activationBlockNumber <= activationBlockNumber) {
                return configs[i];
            }
        }
        revert("unreachable");
    }

    function getCurrentActiveConfig() public view returns (SetConfig memory) {
        return getActiveConfig(uint64(block.number));
    }
}
