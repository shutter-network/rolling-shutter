// SPDX-License-Identifier: MIT

pragma solidity ^0.8.9;

import "hardhat/console.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

struct AddrSet {
    address[] addrs;
}

/**
@title AddrsSeq manages a sequence of a list of addresses

@dev This sequence of addresses is used to store the list of keypers, decryptors or the single
collator at different points in time. Contrary to what we've done in main chain shutter, we just
store the list of addresses here without any information about its validity period. A single list
can be referenced by its index. New lists can be created by calling `add` multiple times and
appended with a final call to `append`.
*/
contract AddrsSeq is Ownable {
    event Appended(uint64 n);
    AddrSet[] private seq;

    constructor() {
        _pushEmptyList();
    }

    function _pushEmptyList() internal {
        seq.push(AddrSet({addrs: new address[](0)}));
    }

    /**
       @notice add adds new addresses to the current list of addresses. This list can be appended
       to the sequence of lists by calling `append`.
     */
    function add(address[] calldata newAddrs) public onlyOwner {
        uint64 idx = uint64(seq.length - 1);

        for (uint64 i = 0; i < newAddrs.length; i++) {
            seq[idx].addrs.push(newAddrs[i]);
        }
    }

    /**
       @notice count returns the number of appended lists
     */
    function count() public view returns (uint64) {
        return uint64(seq.length) - 1;
    }

    /**
     @notice countNth returns the number of addresses stored in the list at index n
    */
    function countNth(uint64 n) public view returns (uint64) {
        require(n < count(), "AddrsSeq.countNth: n out of range");
        return uint64(seq[n].addrs.length);
    }

    /**
       @notice append appends the current list of addresses added with add to the sequence.
     */
    function append() public onlyOwner {
        require(
            seq.length < type(uint64).max - 1,
            "AddrsSeq.append: seq exceeeds limit"
        );
        emit Appended(uint64(seq.length) - 1);
        _pushEmptyList();
    }

    /**
       @notice at returns the address at index i of the list at index n
     */
    function at(uint64 n, uint64 i) public view returns (address) {
        require(n < count(), "AddrsSeq.at: n out of range");
        require(i < seq[n].addrs.length, "AddrsSeq.at: i out of range");
        return seq[n].addrs[i];
    }
}
