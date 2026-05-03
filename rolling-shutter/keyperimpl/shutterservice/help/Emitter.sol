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

    event SingleIdx(uint256 indexed one, bytes two, uint256 three);

    event TransferLike(address indexed from, address indexed to, uint256 value);
    event StaticArgs(
        address user,
        bool ok,
        bytes4 sig,
        bytes32 tag,
        uint256 amount
    );
    event DynamicArgs(string note, bytes blob, uint256[] nums);
    event IndexedDynamic(string indexed note, bytes indexed blob);

    function emitDynamic(
        string memory note,
        bytes memory blob,
        uint256[] memory nums
    ) external {
        emit DynamicArgs(note, blob, nums);
    }

    function emitTransferLike(
        address from,
        address to,
        uint256 value
    ) external {
        emit TransferLike(from, to, value);
    }

    function emitStaticSample() external {
        emit StaticArgs(
            0x1111111111111111111111111111111111111111,
            true,
            0xdeadbeef,
            bytes32("tag"),
            42
        );
    }

    function emitStaticCustom(
        address user,
        bool ok,
        bytes4 sig,
        bytes32 tag,
        uint256 amount
    ) external {
        emit StaticArgs(user, ok, sig, tag, amount);
    }

    function emitDynamicSample() external {
        uint256[] memory nums = new uint256[](2);
        nums[0] = 1;
        nums[1] = 2;
        emit DynamicArgs("hello", hex"beef", nums);
    }

    function emitIndexedDynamicSample() external {
        emit IndexedDynamic("hello", hex"beef");
    }

    function emitSingleIdx(
        uint256 one,
        bytes memory two,
        uint256 three
    ) public {
        emit SingleIdx(one, two, three);
    }

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
