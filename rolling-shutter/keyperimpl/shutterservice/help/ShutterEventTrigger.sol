// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./Ownable.sol";

/**
 * @title ShutterRegistry
 * @dev A contract for managing the registration of identities with timestamps, ensuring unique and future-dated registrations.
 * Inherits from OpenZeppelin's Ownable contract to enable ownership-based access control.
 */
contract ShutterRegistry is Ownable {
    // Custom error for when an identity is already registered.
    error AlreadyRegistered();

    // Custom error for when a provided timestamp is in the past.
    error TimestampInThePast();

    // Custom error for when a identityPrefix provided is empty.
    error InvalidIdentityPrefix();

    // Custom error for when the registration ttl is lower than already registered.
    error TTLTooShort();

    struct EventTriggerRegistration {
        uint64 eon;
        uint64 ttl;
        bytes32 triggerDefinitionHash;
    }
    /**
     * @dev Mapping to store registration data for each identity.
     *      The identity is represented as a `bytes32` hash and mapped to the EventTriggerRegistration.
     */

    mapping(bytes32 identity => EventTriggerRegistration) public registrations;

    /**
     * @dev Emitted when a new event trigger identity is successfully registered.
     * @param eon The eon associated with the identity.
     * @param identityPrefix The raw prefix input used to derive the registered identity hash.
     * @param sender The address of the account that performed the registration.
     * @param triggerDefinition The eventTriggerDefinition associated with the registered identity.
     * @param ttl The blockNumber after which the eventTrigger can be ignored by keypers.
     */
    event EventTriggerRegistered(
        uint64 indexed eon,
        bytes32 identityPrefix,
        address sender,
        bytes[] triggerDefinition,
        uint64 ttl
    );

    /**
     * @dev Initializes the contract and assigns ownership to the deployer.
     */
    constructor() Ownable(msg.sender) {}

    /**
     * @notice Registers a new identity with a specified eventTriggerDefinition and eon.
     * @dev The identity is derived by hashing the provided `identityPrefix` concatenated with the sender's address.
     * @param eon The eon associated with the identity.
     * @param identityPrefix The input used to derive the identity hash.
     * @param triggerDefinition The eventTriggerDefinition.
     * @param ttl A block number in the future after which the trigger can be ignored by keypers.
     * @custom:requirements
     * - The identity must not already be registered.
     * - The provided ttl block number must not be in the past.
     */
    function register(
        uint64 eon,
        bytes32 identityPrefix,
        bytes[] memory triggerDefinition,
        uint64 ttl
    ) external {
        // Ensure the timestamp is not in the past.
        require(ttl >= block.number, TimestampInThePast());

        // Ensure identityPrefix passed in correct.
        require(identityPrefix != bytes32(0), InvalidIdentityPrefix());

        // Generate the identity hash from the provided prefix and the sender's address.
        bytes32 identity = keccak256(
            abi.encodePacked(identityPrefix, msg.sender)
        );
        EventTriggerRegistration storage registrationData = registrations[
            identity
        ];
        // Ensure no one is trying to decrease ttl.
        require(registrationData.ttl < ttl, TTLTooShort());

        if (registrationData.ttl != 0) {
            require(
                keccak256(abi.encode(triggerDefinition)) ==
                    registrationData.triggerDefinitionHash,
                AlreadyRegistered()
            );
        }

        // store the (maybe renewed) registration data;
        registrationData.ttl = ttl;
        registrationData.eon = eon;
        registrationData.triggerDefinitionHash = keccak256(
            abi.encode(triggerDefinition)
        );

        // Emit the EventTriggerRegistered event.
        emit EventTriggerRegistered(
            eon,
            identityPrefix,
            msg.sender,
            triggerDefinition,
            ttl
        );
    }
}
