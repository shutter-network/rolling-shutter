// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract Emitter {
    event Two(uint256 indexed newValue, uint256 value);
    event Four(
        uint256 indexed one,
        uint256 indexed two,
        uint256 indexed three,
        bytes four
    );
    event Five(
        uint256 indexed one,
        uint256 indexed two,
        uint256 indexed three,
        bytes four,
        bytes five
    );
    event Six(
        uint256 indexed one,
        string indexed two,
        address indexed three,
        bytes four,
        uint256 five,
        bytes six
    );

    function emitTwo(uint256 value) public {
        emit Two(value, 5);
    }

    function emitFour(
        uint256 one,
        uint256 two,
        uint256 three,
        bytes memory four
    ) public {
        emit Four(one, two, three, four);
    }

    function emitFive(
        uint256 one,
        uint256 two,
        uint256 three,
        bytes memory four,
        bytes memory five
    ) public {
        emit Five(one, two, three, four, five);
    }

    function emitSix(
        uint256 one,
        string memory two,
        address three,
        bytes memory four,
        uint256 five,
        bytes memory six
    ) public {
        emit Six(one, two, three, four, five, six);
    }
}
