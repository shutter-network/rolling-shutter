// SPDX-License-Identifier: MIT

pragma solidity =0.8.22;

import "./AddrsSeq.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

struct KeypersConfig {
    uint64 activationBlockNumber;
    uint64 setIndex;
    uint64 threshold;
}

contract KeypersConfigsList is Ownable {
    KeypersConfig[] public keypersConfigs;
    AddrsSeq public addrsSeq;

    event NewConfig(
        uint64 activationBlockNumber,
        uint64 keyperSetIndex,
        uint64 keyperConfigIndex,
        uint64 threshold
    );

    constructor(AddrsSeq _addrsSeq) Ownable(msg.sender) {
        addrsSeq = _addrsSeq;
        require(
            addrsSeq.countNth(0) == 0,
            "AddrsSeq must have empty list at index 0"
        );
        keypersConfigs.push(
            KeypersConfig({activationBlockNumber: 0, setIndex: 0, threshold: 0})
        );
        emit NewConfig({
            activationBlockNumber: 0,
            keyperSetIndex: 0,
            keyperConfigIndex: 0,
            threshold: 0
        });
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
        uint64 numKeypers = addrsSeq.countNth(config.setIndex);
        if (numKeypers == 0) {
            require(
                config.threshold == 0,
                "Threshold must be zero if keyper set is empty"
            );
        } else {
            require(config.threshold >= 1, "Threshold must be at least one");
            require(
                config.threshold <= numKeypers,
                "Threshold must not exceed keyper set size"
            );
        }

        keypersConfigs.push(config);
        emit NewConfig({
            activationBlockNumber: config.activationBlockNumber,
            keyperSetIndex: config.setIndex,
            keyperConfigIndex: uint64(keypersConfigs.length) - 1,
            threshold: config.threshold
        });
    }

    function getActiveConfig(
        uint64 activationBlockNumber
    ) public view returns (KeypersConfig memory) {
        for (uint256 i = keypersConfigs.length - 1; true; i--) {
            if (
                keypersConfigs[i].activationBlockNumber <= activationBlockNumber
            ) {
                return keypersConfigs[i];
            }
        }
        revert("unreachable");
    }
}
