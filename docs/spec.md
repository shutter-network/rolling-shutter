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

Shuttermint is a Tendermint blockchain. The keypers use it to run the DKG setup
procedures on it.

Since Shuttermint cannot access or verify external information on its own, the
keypers enter it in the form of transactions and vote on it in order to accept
it. In particular, this applies to the keyper sets which are defined on the
mainchain.

The validators of Shuttermint are the keypers themselves.

### Initialization

### Messages

Shuttermint receives transactions as base64 URL-encoded byte strings.
Transactions consist of a signature followed by the encoded message. The
signature is a 65-byte ECDSA signature created by the sender over the hash of
the encoded message. Messages are encoded according to a protobuf using the
following wrapper:

```
message MessageWithNonce {
        Message msg = 1;
        bytes chain_id = 2;
        uint64 random_nonce = 3;
}

message Message {
        oneof payload {
                BatchConfig batch_config = 4;
                BlockSeen block_seen = 14;
                CheckIn check_in = 7;

                PolyEval poly_eval = 9;
                PolyCommitment poly_commitment = 10;
                Accusation accusation = 11;
                Apology apology = 12;

                EonStartVote eon_start_vote = 13;
        }
}
```

See below for the definitions of the payload types.

Shuttermint rejects transactions that do not follow the above format.

#### Chain Id

Shuttermint has a chain id. Messages whose chain id differs are rejected.

#### Nonce Handling And Spam Protection

Shuttermint accepts transactions only by senders who are part of a keyper set in
any accepted config. Transactions by other senders are rejected.

Messages specify a nonce. Shuttermint tracks the nonces used by each sender.
Messages that carry a nonce that has already been used before by the sender are
rejected.

Shuttermint also tracks the number of transactions each sender has sent per
block. If a sender has reached a limit defined by a constant, any additional
transaction from that sender is rejected.

#### Message Types

Shuttermint is driven by messages sent by keypers. The following describes the
message format as well as how they progress the state.

##### Check In

The check in message is used by keypers to register their identity on the chain.

```
message CheckIn {
        bytes validator_public_key = 1;
        bytes encryption_public_key = 2;
}
```

Shuttermint rejects the message

- if the keyper is already registered.
- `validator_public_key` is not a valid 32-byte compressed ecies public key
- `encryption_public_key` is not a valid compressed ecies public key

Otherwise, Shuttermint stores the validator public key in its state (accessible
by the keyper's address) and emits the following event:

```
type CheckIn struct {
	Height              int64
	Sender              common.Address
	EncryptionPublicKey *ecies.PublicKey
}
```

with `Height` being the current Shuttermint block height, `Sender` the address
of the keyper checking in, and `EncryptionPublicKey` the parsed
`encryption_public_key` from the message.

##### Batch Config

The batch config message is sent to notify Shuttermint of a new keyper
configuration on the mainchain.

```
message BatchConfig {
        uint64 activation_block_number = 1;
        repeated bytes keypers = 2;
        uint64 threshold = 3 ;
        uint64 config_index = 5;
        bool started = 6;
        bool validatorsUpdated = 7;
}
```

Shuttermint rejects the message if

- `keypers` is empty
- one of the entries in `keypers` is of a size different than 20 bytes
- `keypers` contains duplicates
- `threshold` is zero or greater than the number of keypers
- the last config accepted to Shuttermint has a greater activation block number
  or greater or equal config index (TODO: allow for out-of-order configs)
- the sender is not a member of the keyper set defined in the last accepted
  config (as only those are allowed to change the config)

Shuttermint records a vote by the sender for the config. If a vote has already
been recorded, the message is rejected.

If at least the threshold of the last accepted config has voted for the received
config, the config is added as the now last accepted config. All votes
(including for other configs) are reset. In addition, Shuttermint will start a
new DKG process for the keyper set defined in the config.

If the config is accepted, the following event is emitted:

```
type EonStarted struct {
	Height                int64
	Eon                   uint64
	ActivationBlockNumber uint64
	ConfigIndex           uint64
}
```

where `Height` is the current block height, `Eon` is a counter value uniquely
identifying the started DKG process, and `ActivationBlockNumber` as well as
`ConfigIndex` are taken from the submitted and accepted batch config.

##### Block Seen

The block seen message is sent to notify Shuttermint of mainchain blocks that
have passed.

```
message BlockSeen {
        uint64 block_number = 1;
}
```

Shuttermint keeps track of the greatest block seen by each sender. Shuttermint
will update this number if `block_number` is greater than the current value for
the sender of the message.

Block seen messages do not emit events.

##### Eon Start Vote

Keypers send eon start votes in order to coordinate on the start of a new DKG
process.

```
message EonStartVote {
        uint64 activation_block_number = 1;
}
```

TODO: add config index?

Shuttermint rejects the message if the sender is not part of the keyper set of
the config identified by `activation_block_number` (note that due to the guard
element with activation block number 0, there will always be matching config).

Shuttermint records the vote by the sender for the given activation block
number. If the sender has already voted for a different activation block number,
their vote is updated.

If at least `threshold` (defined in the config identified by the activation
block number) votes have been recorded for the same activation block number, a
new DKG process is started and the following event is emitted:

```
type EonStarted struct {
	Height                int64
	Eon                   uint64
	ActivationBlockNumber uint64
	ConfigIndex           uint64
}
```

where `Height` is the current block height, `Eon` is a counter value uniquely
identifying the started DKG process, and `ActivationBlockNumber` as well as
`ConfigIndex` are given by the config for which the key shall be generated.

### Post Block Processing

## Collator

## Keyper

## Rollup State Execution
