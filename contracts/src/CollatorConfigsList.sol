// SPDX-License-Identifier: MIT

pragma solidity =0.8.9;

import "./AddrsSeq.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

struct CollatorConfig {
    uint64 activationBlockNumber;
    uint64 setIndex;
}

contract CollatorConfigsList is Ownable {
    CollatorConfig[] public collatorConfigs;
    AddrsSeq public addrsSeq;

    event NewConfig(uint64 activationBlockNumber, uint64 index);

    constructor(AddrsSeq _addrsSeq) {
        addrsSeq = _addrsSeq;
        require(
            addrsSeq.countNth(0) == 0,
            "AddrsSeq must have empty list at index 0"
        );
        collatorConfigs.push(
            CollatorConfig({activationBlockNumber: 0, setIndex: 0})
        );
        emit NewConfig(0, 0);
    }

    function addNewCfg(CollatorConfig calldata config) public onlyOwner {
        require(
            addrsSeq.count() > config.setIndex,
            "No appended set in seq corresponding to config's set index"
        );
        require(
            collatorConfigs[collatorConfigs.length - 1].activationBlockNumber <=
                config.activationBlockNumber,
            "Cannot add new set with lower block number than previous"
        );
        require(
            block.number <= config.activationBlockNumber,
            "Cannot add new set with past block number"
        );

        collatorConfigs.push(config);
        emit NewConfig(config.activationBlockNumber, config.setIndex);
    }

    function getActiveConfig(uint64 activationBlockNumber)
        public
        view
        returns (CollatorConfig memory)
    {
        for (uint256 i = collatorConfigs.length - 1; true; i--) {
            if (
                collatorConfigs[i].activationBlockNumber <=
                activationBlockNumber
            ) {
                return collatorConfigs[i];
            }
        }
        revert("unreachable");
    }

    function getCurrentActiveConfig()
        public
        view
        returns (CollatorConfig memory)
    {
        return getActiveConfig(uint64(block.number));
    }
}
