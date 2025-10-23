# EventTriggerDefinition Format Specification

This specification provides the format and encoding rules necessary to implement
tooling that generates `EventTriggerDefinition` instances programmatically.
These
[`triggerDefinition`](https://github.com/shutter-network/contracts/blob/8b0051b13875a7450811c4634e72f2edad7f5016/src/shutter-service/ShutterEventTriggerRegistry.sol#L38)s
are used to register event based decryption triggers. Keypers that take part in
a keyper set that has these enabled, will watch the blockchain for events
matching these definitions. The registration is done through a contract specific
to the keyperset. An example
[can be seen here](https://github.com/shutter-network/contracts/blob/8b0051b13875a7450811c4634e72f2edad7f5016/src/shutter-service/ShutterEventTriggerRegistry.sol).

## Elements / Types

### EventTriggerDefinition

An `EventTriggerDefinition` is a structured data object that combines:

- A reference to a specific smart contract (by address)
- A set of log predicates that filter which events from that contract should
  trigger the release of the decryption key.

Reference:
[EventTriggerDefinition Type](https://github.com/shutter-network/rolling-shutter/blob/42f562532acfc4f89f630d3de809fc4451636ab2/rolling-shutter/keyperimpl/shutterservice/eventtrigger.go#L21-L25)

### LogPredicate

A `LogPredicate` pairs a reference to a specific value within an event log with
a predicate that must be satisfied for the trigger to fire.

Reference:
[LogPredicate Type](https://github.com/shutter-network/rolling-shutter/blob/42f562532acfc4f89f630d3de809fc4451636ab2/rolling-shutter/keyperimpl/shutterservice/eventtrigger.go#L27-L32)

### LogValueRef

A `LogValueRef` identifies which value within an event log should be evaluated.
It can reference:

- **Topics (Offsets 0-3)**: The indexed parameters of the event log. A topic is
  referenced by its index (0-3), where offset 0 is topic[0], etc.
- **Data (Offsets 4+)**: The non-indexed data section of the log. Offset
  values >= 4 refer to 32-byte words in the log data, where offset 4 is the
  first word (bytes 0-31), offset 5 is the second word (bytes 32-63), etc.

Reference:
[LogValueRef Type and Documentation](https://github.com/shutter-network/rolling-shutter/blob/42f562532acfc4f89f630d3de809fc4451636ab2/rolling-shutter/keyperimpl/shutterservice/eventtrigger.go#L34-L43)

### ValuePredicate

A `ValuePredicate` defines the condition that must be satisfied on a referenced
log value. It consists of an operation (`Op`) and a set of arguments. The type
and number of arguments required depend on the operation.

Reference:
[ValuePredicate Type](https://github.com/shutter-network/rolling-shutter/blob/42f562532acfc4f89f630d3de809fc4451636ab2/rolling-shutter/keyperimpl/shutterservice/eventtrigger.go#L45-L53)

## Encoding Format

### Binary Encoding

`EventTriggerDefinition` values are encoded using RLP (Recursive Length Prefix)
encoding with a version byte prefix:

```
EventTriggerDefinitionBytes := {
    version: uint8                    // Current version is 0x01
    rlp_encoded_content: []byte       // RLP-encoded EventTriggerDefinition
}
```

The version byte (`0x01`) allows for future format changes. After the version
byte, the remaining bytes are RLP-encoded content representing the
`EventTriggerDefinition` structure.

Reference:
[MarshalBytes and UnmarshalBytes Implementation](https://github.com/shutter-network/rolling-shutter/blob/42f562532acfc4f89f630d3de809fc4451636ab2/rolling-shutter/keyperimpl/shutterservice/eventtrigger.go#L58-L86)

### Versioning

The current version is **0x01** as defined by the `Version` constant.

### RLP Encoding Details

The RLP encoding format for the core structures is:

```
EventTriggerDefinition := {
    contract: Address,
    logPredicates: [LogPredicate, ...]
}

LogPredicate := {
    logValueRef: LogValueRef,
    valuePredicate: ValuePredicate
}

LogValueRef := Offset                    // if Length == 1
LogValueRef := [Offset, Length]          // if Length > 1

ValuePredicate := {
    op: uint64,
    intArgs: [BigInt, ...],
    byteArgs: [Bytes, ...]
}
```

Reference:
[RLP Encoding Implementation](https://github.com/shutter-network/rolling-shutter/blob/42f562532acfc4f89f630d3de809fc4451636ab2/rolling-shutter/keyperimpl/shutterservice/eventtrigger.go#L211-L267)

## Operators

The `Op` type defines the operations that can be performed when evaluating a
value predicate.

### Supported Operators

- **`Uint...` (0-4) Operators**: Unsigned integer (`uint256` / `big.Int`)
  comparisons
  - Argument: 1 unsigned integer
  - **`UintLt` (0)**
    - Returns true if `value < argument`
  - **`UintLte` (1)**
    - Returns true if `value <= argument`
  - **`UintEq` (2)**
    - Returns true if `value == argument`
  - **`UintGt` (3)**
    - Returns true if `value > argument`
  - **`UintGte` (4)**
    - Returns true if `value >= argument`
- **`BytesEq` (5)**: Byte sequence equality comparison
  - Arguments: 1 byte sequence (exactly matching the value size)
  - Returns true if value == argument (byte-by-byte comparison)

Reference:
[Operator Constants](https://github.com/shutter-network/rolling-shutter/blob/42f562532acfc4f89f630d3de809fc4451636ab2/rolling-shutter/keyperimpl/shutterservice/eventtrigger.go#L298-L305)

Note, that different operators require different numbers and types of arguments:

- **Uint...-operators**: 1 integer argument, 0 byte arguments
- **BytesEq**: 0 integer arguments, 1 byte argument

## Matching Logs Against Definitions

In the keyper implementation, the `Match` method is used to check if a given
Ethereum log satisfies the `EventTriggerDefinition`:

1. The log's contract address must match the definition's contract address
2. The log must satisfy **all** log predicates (logical `AND` of all conditions)

Reference:
[Match Method](https://github.com/shutter-network/rolling-shutter/blob/42f562532acfc4f89f630d3de809fc4451636ab2/rolling-shutter/keyperimpl/shutterservice/eventtrigger.go#L159-L176)

### LogValueRef.GetValue

To retrieve a value from a log:

- **Topics (Offset < 4)**: Returns the log's topic at index Offset as a 32-byte
  value
- **Data (Offset >= 4)**: Extracts a slice of 32-byte words from the log's data,
  starting at word index (Offset - 4). If the slice extends beyond the log's
  data, it is zero-padded on the right to the expected length.

Reference:
[GetValue Method](https://github.com/shutter-network/rolling-shutter/blob/main/rolling-shutter/keyperimpl/shutterservice/eventtrigger.go#L210-L228)

## Generating EventTriggerDefinition from EVM Contract ABI

To generate an `EventTriggerDefinition` from an EVM contract ABI and a set of
predicates:

### Input Requirements

1. **Contract Address**: The Ethereum address where the contract is deployed
2. **Contract ABI**: Standard Solidity JSON ABI format
3. **Target Event Name**: The name of the event to trigger on
4. **Predicate Specifications**: List of conditions, each specifying:
   - Parameter name (e.g., "to", "amount") or parameter index
   - Whether it's indexed or not (helps determine offset calculation)
   - Operation to apply (UintLt, UintEq, BytesEq, etc.)
   - Argument value(s)

### Algorithm Overview

```
function generateEventTriggerDefinition(
    contractAddress: Address,
    abi: ContractABI,
    eventName: string,
    predicateSpecs: PredicateSpec[]
): EventTriggerDefinition {

    // 1. Find the event in the ABI
    event = findEventInABI(abi, eventName)
    if (!event) {
        throw EventNotFoundError(eventName)
    }

    // 2. Build log predicates from predicate specifications
    logPredicates = []

    for spec in predicateSpecs {
        // Find the parameter in the event
        param = findParameterInEvent(event, spec.parameterName)
        if (!param) {
            throw ParameterNotFoundError(spec.parameterName)
        }

        // Calculate the LogValueRef offset
        if (param.indexed) {
            // Topics are at offsets 0-3
            offset = calculateTopicIndex(event, param)
            length = 1
        } else {
            // Data starts at offset 4
            offset = 4 + calculateDataWordOffset(event, param)
            length = calculateDataWordLength(param.type)
        }

        logValueRef = LogValueRef{
            Offset: offset,
            Length: length
        }

        // Create the ValuePredicate
        valuePredicate = createValuePredicate(
            spec.operation,
            spec.value,
            param.type
        )

        logPredicates.append(LogPredicate{
            LogValueRef: logValueRef,
            ValuePredicate: valuePredicate
        })
    }

    // 3. Assemble and return
    return EventTriggerDefinition{
        Contract: contractAddress,
        LogPredicates: logPredicates
    }
}
```

### Calculating Offsets

**For Indexed Parameters (Topics)**:

Topics are indexed parameters in Solidity events and are filtered by topic
values in the EVM log. The offset corresponds to the topic index:

- Topic 0 is always the event signature hash and is implicit (offset 0 if
  included)
- Additional topics (offset 1, 2, 3) correspond to indexed parameters

**For Non-Indexed Parameters (Data)**:

Non-indexed parameters are stored in the log's data field. They are packed as
32-byte words:

- Word 0 (bytes 0-31) corresponds to offset 4
- Word 1 (bytes 32-63) corresponds to offset 5
- And so on...

For types smaller than 32 bytes, they are right-padded (for fixed-size types) or
stored in a single word.

## Examples

### Example 1: ERC20 Transfer Trigger

Trigger when a specific address transfers tokens:

**Event ABI:**

```json
{
  "name": "Transfer",
  "inputs": [
    { "name": "from", "type": "address", "indexed": true },
    { "name": "to", "type": "address", "indexed": true },
    { "name": "value", "type": "uint256", "indexed": false }
  ]
}
```

**Predicate Specification:**

- Parameter "from" equals 0x742d35Cc6634C0532925a3b844Bc9e7595f1bEb

**Generated EventTriggerDefinition (JSON):**

```json
{
  "version": 1,
  "contract": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
  "logPredicates": [
    {
      "logValueRef": {
        "offset": 1,
        "length": 1
      },
      "valuePredicate": {
        "op": 5,
        "intArgs": [],
        "byteArgs": [
          "0x000000000000000000000000742d35cc6634c0532925a3b844bc9e7595f1beb"
        ]
      }
    }
  ]
}
```

### Example 2: Multi-Condition Trigger

Trigger when a swap occurs with minimum token amount:

**Event ABI (Uniswap V3 SwapRouter):**

```json
{
  "name": "Swap",
  "inputs": [
    { "name": "sender", "type": "address", "indexed": true },
    { "name": "recipient", "type": "address", "indexed": true },
    { "name": "amount0Delta", "type": "int256", "indexed": false },
    { "name": "amount1Delta", "type": "int256", "indexed": false }
  ]
}
```

**Predicate Specifications:**

- Parameter "amount1Delta" >= 1000000000000000000 (1 token with 18 decimals)
- Parameter "sender" equals 0x742d35Cc6634C0532925a3b844Bc9e7595f1bEb

**Generated EventTriggerDefinition (JSON):**

```json
{
  "version": 1,
  "contract": "0x1111111254fb6d44bac0bed2854e76f90643097d",
  "logPredicates": [
    {
      "logValueRef": {
        "offset": 1,
        "length": 1
      },
      "valuePredicate": {
        "op": 5,
        "intArgs": [],
        "byteArgs": [
          "0x000000000000000000000000742d35cc6634c0532925a3b844bc9e7595f1beb"
        ]
      }
    },
    {
      "logValueRef": {
        "offset": 5,
        "length": 1
      },
      "valuePredicate": {
        "op": 4,
        "intArgs": ["1000000000000000000"],
        "byteArgs": []
      }
    }
  ]
}
```
