// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

contract SimpleStorage {
    uint256 private value;
    event ValueChanged(uint256 newValue);

    constructor(uint256 initialValue) {
        value = initialValue;
    }

    function setValue(uint256 newValue) public {
        value = newValue;
        emit ValueChanged(newValue);
    }

    function getValue() public view returns (uint256) {
        return value;
    }

    // Function that will consume different amounts of gas
    function expensiveOperation(uint256 iterations) public {
        uint256 result = 0;
        for(uint256 i = 0; i < iterations; i++) {
            result += i;
        }
        value = result;
        emit ValueChanged(result);
    }
}
