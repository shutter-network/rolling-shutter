// SPDX-License-Identifier: MIT

pragma solidity =0.8.9;

contract BatchCounter {
    error CallerNotZeroAddress();

    event NewBatchIndex(uint64 oldIndex, uint64 newIndex);
    uint64 public batchIndex;

    function increment() external {
        if (msg.sender != address(0)) {
            revert CallerNotZeroAddress();
        }
        emit NewBatchIndex(batchIndex, batchIndex + 1);
        batchIndex += 1;
    }

    /// @notice intended for initialization
    function set(uint64 newBatchIndex) external {
        if (msg.sender != address(0)) {
            revert CallerNotZeroAddress();
        }
        emit NewBatchIndex(batchIndex, newBatchIndex);
        batchIndex = newBatchIndex;
    }
}
