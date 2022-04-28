// SPDX-License-Identifier: MIT

pragma solidity =0.8.9;

import "@openzeppelin/contracts/access/Ownable.sol";

contract BatchCounter is Ownable {
    event NewBatchIndex(uint64 oldIndex, uint64 newIndex);
    uint64 public batchIndex;

    function increment() external onlyOwner {
        emit NewBatchIndex(batchIndex, batchIndex + 1);
        batchIndex += 1;
    }

    /// @notice intended for initialization
    function set(uint64 newBatchIndex) external onlyOwner {
        emit NewBatchIndex(batchIndex, newBatchIndex);
        batchIndex = newBatchIndex;
    }
}
