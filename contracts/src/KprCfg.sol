pragma solidity =0.8.9;

import "./AddrsSeq.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

struct KprCfg {
    uint64 blkNbr;
    uint64 index;
}

contract KprCfgs is Ownable {
    KprCfg[] public kprCfgs;
    AddrsSeq public addrsSeq;

    event NewCfg(uint64 blkNbr, uint64 i);

    constructor(AddrsSeq a) {
        addrsSeq = a;
        kprCfgs.push(KprCfg({blkNbr: 0, index: 0}));
        emit NewCfg(0, 0);
    }

    function addNewCfg(KprCfg calldata cfg) public onlyOwner {
        require(
            addrsSeq.count() >= cfg.index,
            "No appended set in seq corresponding to index"
        );
        require(
            kprCfgs[kprCfgs.length - 1].blkNbr <= cfg.blkNbr,
            "Cannot add new set with lower block number than previous"
        );

        kprCfgs.push(cfg);
        emit NewCfg(cfg.blkNbr, cfg.index);
    }

    function getActiveCfg(uint64 blkNbr) public view returns (KprCfg memory) {
        for (uint256 i = kprCfgs.length - 1; true; i--) {
            if (kprCfgs[i].blkNbr <= blkNbr) {
                return kprCfgs[i];
            }
        }
        revert("unreachable");
    }

    function getCurrentActiveCfg() public view returns (KprCfg memory) {
        return getActiveCfg(uint64(block.number));
    }
}
