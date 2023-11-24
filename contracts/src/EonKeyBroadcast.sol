// SPDX-License-Identifier: MIT

pragma solidity =0.8.22;

import "@openzeppelin/contracts/access/Ownable.sol";

/// @notice EonKeyStorage allows its owner to publish a sequence of eon keys, each valid from a
/// given activation block number on. New keys can be added at any time, even out of order or for
/// past activation blocks. Anyone can retrieve the key that is valid for a given block number.
contract EonKeyStorage is Ownable {
    event Inserted(uint64 activationBlockNumber, uint64 index, bytes key);
    error NotFound(uint64 blockNumber);

    // Keys are stored as a linked list, sorted from greatest to smallest activation block
    // number. The list has guard elements at the head and tail. The head element with activation
    // block number uint64 max is stored at index 1, the tail lives at index 0 with activation
    // block number 0. The guards allow us to simplify the code, because we always insert between
    // two keys.

    // nextIndex is the index of the key with the next smaller (or equal) activation block
    // number. In case of a tie, keys added later come before older keys, effectively replacing
    // them.
    struct Key {
        uint64 activationBlockNumber;
        uint64 nextIndex;
        bytes key;
    }
    Key[] public keys;

    constructor() Ownable(msg.sender) {
        bytes memory empty;
        _insertKey(type(uint64).min, empty, 1);
        _insertKey(type(uint64).max, empty, 0);
    }

    /// @notice Get the number of keys in the storage.
    function num() external view returns (uint64) {
        return uint64(keys.length) - 2;
    }

    /// @notice Add a new key to the storage.
    /// @param serializedKey The key to insert
    /// @param activationBlockNumber The block number from which on the key shall be used
    function insert(
        bytes memory serializedKey,
        uint64 activationBlockNumber
    ) external onlyOwner {
        uint64 newIndex;
        uint64 index = 1;
        Key memory key = keys[index];

        // Search for the spot to insert the key, i.e., right before the first key with a smaller
        // (or equal) activation block number. Stop searching at the end of the list.
        while (true) {
            Key memory nextKey = keys[key.nextIndex];
            if (nextKey.activationBlockNumber <= activationBlockNumber) {
                // Insert between nextKey and key, udpating key's nextKey pointer to the newly
                // added key.
                newIndex = _insertKey(
                    activationBlockNumber,
                    serializedKey,
                    key.nextIndex
                );
                key.nextIndex = newIndex;
                keys[index] = key;
                return;
            }

            index = key.nextIndex;
            key = nextKey;
        }
    }

    /// @notice Retrieve a key.
    /// @param blockNumber The block number for which the key shall be used. The returned key will
    /// have an activation block number smaller or equal to this block number.
    function get(uint64 blockNumber) external view returns (bytes memory) {
        // Iterate through all keys, starting with the key following the head guard at index 1. The
        // keys are ordered by activation block number, so the first one with activation block
        // number <= block number is the one we're looking for.
        uint64 index = keys[1].nextIndex;
        while (index != 0) {
            Key memory key = keys[index];
            if (key.activationBlockNumber <= blockNumber) {
                return key.key;
            }
            index = key.nextIndex;
        }
        revert NotFound(blockNumber);
    }

    function _insertKey(
        uint64 activationBlockNumber,
        bytes memory key,
        uint64 nextIndex
    ) internal returns (uint64) {
        uint64 index = uint64(keys.length);
        keys.push(
            Key({
                activationBlockNumber: activationBlockNumber,
                key: key,
                nextIndex: nextIndex
            })
        );
        // Do not emit the event for guard elements
        if (index >= 2) {
            emit Inserted({
                activationBlockNumber: activationBlockNumber,
                key: key,
                index: index
            });
        }
        return index;
    }
}
