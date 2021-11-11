// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import "./Registry.sol";
import "./SetConfigsList.sol";

contract DecryptorsConfigsList is SetConfigsList {
    Registry public BLSKeysRegistry;
    Registry public KeySignaturesRegistry;

    constructor(
        AddrsSeq _addrsSeq,
        Registry _BLSKeysRegistry,
        Registry _KeySignaturesRegistry
    ) SetConfigsList(_addrsSeq) {
        require(
            _BLSKeysRegistry.addrsSeq() == _addrsSeq,
            "AddrsSeq of _BLSKeysRegistry must be _addrsSeq"
        );
        require(
            _KeySignaturesRegistry.addrsSeq() == _addrsSeq,
            "AddrsSeq of _KeySignaturesRegistry must be _addrsSeq"
        );
        require(
            _KeySignaturesRegistry != _BLSKeysRegistry,
            "The two used registries must be different"
        );

        BLSKeysRegistry = _BLSKeysRegistry;
        KeySignaturesRegistry = _KeySignaturesRegistry;
    }
}
