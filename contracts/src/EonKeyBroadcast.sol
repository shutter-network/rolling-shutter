// SPDX-License-Identifier: MIT

pragma solidity =0.8.9;

import "@openzeppelin/contracts/access/Ownable.sol";

/// @notice EonKeyStorage allows its owner to publish a sequence of eon keys, each valid from a
/// given activation block number on. New keys can be added at any time, even out of order or for
/// past activation blocks. Anyone can retrieve the key that is valid for a given block number.
contract EonKeyStorage is Ownable {
    event Inserted(uint64 activationBlockNumber, uint64 index, bytes key);
    error NotFound(uint64 blockNumber);

    // Keys are stored as a linked list, sorted from greatest to smallest activation block number.
    // The key with greatest activation block number is at index firstKeyIndex. nextIndex is the
    // index of the key with the next smaller (or equal) activation block number. In case of a tie,
    // keys added later come before older keys, effectively replacing them.
    struct Key {
        uint64 activationBlockNumber;
        uint64 nextIndex;
        bytes key;
    }
    Key[] public keys;
    uint64 public firstKeyIndex;

    /// @notice Get the number of keys in the storage.
    function num() external view returns (uint64) {
        return uint64(keys.length);
    }

    /// @notice Add a new key to the storage.
    /// @param serializedKey The key to insert
    /// @param activationBlockNumber The block number from which on the key shall be used
    function insert(bytes memory serializedKey, uint64 activationBlockNumber)
        external
        onlyOwner
    {
        // If it's the first key, simply insert it at the beginning. Set nextIndex to itself
        // to mark the end of the list. firstKeyIndex is already 0, so no need to update it.
        if (keys.length == 0) {
            _insertKey(activationBlockNumber, serializedKey, 0);
            return;
        }

        uint64 newIndex;
        uint64 index = firstKeyIndex;
        Key memory key = keys[index];

        // If the new key has a greater (or equal) activation block number than the first key,
        // make the new key the new head of the list.
        if (key.activationBlockNumber <= activationBlockNumber) {
            newIndex = _insertKey(
                activationBlockNumber,
                serializedKey,
                firstKeyIndex
            );
            firstKeyIndex = newIndex;
            return;
        }

        // Search for the spot to insert the key, i.e., right before the first key with a smaller
        // (or equal) activation block number. Stop searching at the end of the list.
        while (index != key.nextIndex) {
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

        // We've reached the first key and all keys have greater activation block numbers. Thus, we
        // have to insert at the end of the list, pointing to ourselves.
        newIndex = _insertKey(
            activationBlockNumber,
            serializedKey,
            uint64(keys.length)
        );
        key.nextIndex = newIndex;
        keys[index] = key;
        return;
    }

    /// @notice Retrieve a key.
    /// @param blockNumber The block number for which the key shall be used. The returned key will
    /// have an activation block number smaller or equal to this block number.
    function get(uint64 blockNumber) external view returns (bytes memory) {
        if (keys.length == 0) {
            revert NotFound(blockNumber);
        }

        // Iterate through all keys, starting with firstKeyIndex. The keys are ordered by activation
        // block number, so the first one with activation block number <= block number is the one
        // we're looking for.
        uint64 index = firstKeyIndex;
        while (true) {
            Key memory key = keys[index];
            if (key.activationBlockNumber <= blockNumber) {
                return key.key;
            }
            if (index == key.nextIndex) {
                // Only the key with smallest activation block number is allowed to point at
                // itself. Thus, we reached the first key and didn't find anything.
                break;
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
        emit Inserted({
            activationBlockNumber: activationBlockNumber,
            key: key,
            index: index
        });
        return index;
    }
}
