// SPDX-License-Identifier: MIT

pragma solidity =0.8.9;

import "./SetConfigsList.sol";

contract KeypersConfigsList is SetConfigsList {
    constructor(AddrsSeq _addrsSeq) SetConfigsList(_addrsSeq) {}
}
