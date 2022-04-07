# Shutter Spec

## Glossary

## P2P Network

Keypers and collator communicate with each other using gossipsub. Messages are
defined as protobufs and serialized as such. For each message type there is
exactly one gossipsub topic.

Nodes must reject messages that exceed the size limit of xxx bytes. Nodes must
reject messages that cannot be deserialized according to the protobuf definition
corresponding to the topic the message was received on. Nodes must reject
messages if any message type specific validity condition is not met. Those are
defined below.

All message types have an `instanceId` field. Nodes must reject messages if
their instance id does not match the instance they manage.

### Message Types

#### Decryption Trigger

The collator sends decryption triggers to inform the keypers that they shall
start creating a decryption key.

Topic: `decryptionTrigger`

Subscribed by: Collator, keypers

```
message DecryptionTrigger {
    uint64 instanceID = 1;
    uint64 epochID = 2;
    bytes transactionsHash = 3;
    bytes signature = 4;
}
```

Checks:

- `transactionsHash` must be 32 bytes
- `signature` must be a valid ECDSA signature

#### DecryptionKeyShare

Keypers send decryption key shares upon receiving a corresponding decryption
trigger in order to generate the decryption key.

Topic: `decryptionKeyShare`

Subscribed by: Keypers

```
message DecryptionKeyShare {
    uint64 instanceID = 1;
    uint64 epochID = 2;
    uint64 keyperIndex = 3;
    bytes share = 4;
}
```

Checks:

- `share` must be a valid epoch decryption key share

#### DecryptionKey

Keypers broadcas the decryption key when they managed to aggregate it from the
corresponding shares they have received.

Topic: `decryptionKey`

Subscribed by: Collator, keypers

```
message DecryptionKey {
    uint64 instanceID = 1;
    uint64 epochID = 2;
    bytes key = 3;
}
```

Checks:

- `key` must be a valid epoch decryption key

#### CipherBatch

TODO: remove

Topic: `cipherBatch`

```
message CipherBatch {
    DecryptionTrigger decryption_trigger = 1;
    repeated bytes transactions = 2;
}
```

- `decryption_trigger` must be a valid `DecryptionTrigger` as defined above

#### EonPublicKey

Keypers broadcast the eon public when they have generated a new one to inform
the network about it.

Topic: `eonPublicKey`

Subscribed by: Collator, keypers

```
message EonPublicKey {
    uint64 instanceID = 1;
    bytes publicKey= 2;
    uint64 eon = 3;
}
```

## Smart Contracts

A set of smart contracts is used to configure the system. In case of Rolling
Shutter, they are deployed on the rollup. In case of Snapshot, they can be
deployed on any chain.

### KeypersConfigsList

The keypers configs list defines the keyper sets and threshold parameters for
each block height.

### CollatorConfigsList

The colltator configs list defines the collator for each block height.

### EonKeyStorage

The eon key storage contract stores eon keys and makes them available to
callers.

## Shuttermint

## Collator

## Keyper

## Rollup State Execution
