This documents lists and describes the components that have to be developed or
modified in order to build the off-chain decrypting Rolling Shutter.

### Collator

The collator is a standalone application responsible for creating batches and
submitting them to the rollup for execution. THe collator is operated by the
same entity as the sequencer.

The collator accepts cipher transactions from users via a network interface. At
the end of each epoch, the collator creates a cipher batch based on the
transactions it received. It sends a commitment to the cipher batch as well as
block metadata to the rollup and broadcasts it on a p2p gossip network.

The collator also listens on the network for decryption signatures from the
decryptors. Using these signatures, it creates plaintext batches and submits
them to the sequencer, with metadata matching their commitment.

### Keyper

The keyper application is taken over from On-Chain Shutter. It is modified in
the following ways:

- Releasing the epoch secret key share is triggered by seeing a batch commitment
  on the rollup instead of a certain block number.
- Instead of computing the epoch secret key, decrypting batches, and voting on
  the result, only the epoch secret key share is broadcasted on a public p2p
  gossip network.
- ConfigContract watching?

### Decryptor

The decryptor is a node that decrypts cipher batches. It listens on a p2p
network for cipher batches signed by the collator, as well as decryption keys
from the keypers. Once they have both for a particular epoch, they decrypt the
batch and broadcast it, along with a signature notarising the result.

### Sequencer

The sequencer is modified such that it abides by the batch commitments the
collator gave. It also preferably accepts transactions from the collator and
makes sure to not include transactions between batch commitment and execition.

### Execution Manager

Optimism's execution manager contract is modified such that it handles plaintext
batches. Plaintext batches are executed as follows:

1. Check the batch's epoch index against an epoch counter.
2. Increment the epoch counter.
3. Check the decryption signature.
4. Execute the transactions in the batch one by one.

The executor contract guarantees that if any of the steps fails, the rollup
state is rolled back completely (at least those parts accessible by other
contracts).

The execution manager usually executes other transactions as normal. Only if the
appear between a batch commitment and the end of batch execution, it will be
ignored.

### Collator Slasher

The collator slasher is a contract deployed on mainnet. It slashes the deposit
of the collator for

- creating two batch commitments for the same epoch or
- submitting a batch with metadata that does not match their commitment

Slashing can be triggered by anyone providing an incriminating pair of
commitments, or a single commitment with false metadata, respectively.

### Decryptor Slasher

The decryptor slasher is a contract deployed on mainnet. It slashes the deposits
of decryptors who sign a plaintext batch that does not match the corresponding
cipher batch. Slashing must be triggered by a Kleros court.
