// SPDX-License-Identifier: MIT

pragma solidity ^0.8.9;

import "hardhat/console.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "./AddrsSeq.sol";

/**
   @title Registry allows senders that are contained inside an AddrsSeq to register a single bytes
   value. We use this to let decryptors register their BLS public key.
 */
contract Registry {
    event Registered(uint64 n, uint64 i, address a, bytes data);

    AddrsSeq public addrsSeq;
    mapping(address => bytes) private dataMap;

    constructor(AddrsSeq _addrsSeq) {
        addrsSeq = _addrsSeq;
    }

    /**
       @notice register the given BLS public key for sender. The sender must match the address
       stored at the given coordinates (n,i) in the AddrsSeq contract. Each sender can only
       register a single key.
     */
    function register(
        uint64 n,
        uint64 i,
        bytes memory data
    ) public {
        address a = addrsSeq.at(n, i);
        require(a == msg.sender, "Registry: sender is not allowed");
        require(
            dataMap[msg.sender].length == 0,
            "Registry: sender already registered"
        );
        require(data.length > 0, "Registry: cannot register empty value");
        dataMap[msg.sender] = data;
        console.log("Registered value for sender %s", msg.sender);
        emit Registered(n, i, msg.sender, data);
    }

    /**
     @notice get the registered public key for the given address
    */
    function get(address a) public view returns (bytes memory) {
        return dataMap[a];
    }
}
