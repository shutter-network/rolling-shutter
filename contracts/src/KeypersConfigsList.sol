// SPDX-License-Identifier: MIT

pragma solidity =0.8.9;

import "./AddrsSeq.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

struct KeypersConfig {
    uint64 activationBlockNumber;
    uint64 setIndex;
}

contract KeypersConfigsList is Ownable {
    KeypersConfig[] public keypersConfigs;
    AddrsSeq public addrsSeq;

    event NewConfig(uint64 activationBlockNumber, uint64 index);

    constructor(AddrsSeq _addrsSeq) {
        addrsSeq = _addrsSeq;
        require(
            addrsSeq.countNth(0) == 0,
            "AddrsSeq must have empty list at index 0"
        );
        keypersConfigs.push(
            KeypersConfig({activationBlockNumber: 0, setIndex: 0})
        );
        emit NewConfig(0, 0);
    }

    function addNewCfg(KeypersConfig calldata config) public onlyOwner {
        require(
            addrsSeq.count() > config.setIndex,
            "No appended set in seq corresponding to config's set index"
        );
        require(
            keypersConfigs[keypersConfigs.length - 1].activationBlockNumber <=
                config.activationBlockNumber,
            "Cannot add new set with lower block number than previous"
        );
        require(
            block.number <= config.activationBlockNumber,
            "Cannot add new set with past block number"
        );

        keypersConfigs.push(config);
        emit NewConfig(config.activationBlockNumber, config.setIndex);
    }

    function getActiveConfig(uint64 activationBlockNumber)
        public
        view
        returns (KeypersConfig memory)
    {
        for (uint256 i = keypersConfigs.length - 1; true; i--) {
            if (
                keypersConfigs[i].activationBlockNumber <= activationBlockNumber
            ) {
                return keypersConfigs[i];
            }
        }
        revert("unreachable");
    }

    function getCurrentActiveConfig()
        public
        view
        returns (KeypersConfig memory)
    {
        return getActiveConfig(uint64(block.number));
    }
}
