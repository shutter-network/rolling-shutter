// SPDX-License-Identifier: MIT

pragma solidity =0.8.9;

import "./SetConfigsList.sol";

contract CollatorConfigsList is SetConfigsList {
    constructor(AddrsSeq _addrsSeq) SetConfigsList(_addrsSeq) {}
}
