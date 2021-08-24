### Motivation

- We're having trouble coming up with a complete design for Rolling Shutter.
- The main challenge is the integration into existing rollup implementations in
  a minimally invasive fashion, without breaking censorship or frontrunning
  resistance.
- However, we have a very clear picture of how the solution should look like
  conceptually.
- We therefore came up with a plan for a phased development process in which we
  build major parts of the system first, but postpone decisions about the
  specifics of the integration for later.
- After release of the first one or two phases we will be in a better position
  to talk with rollup teams about support with the integration, they will be
  less busy with launching their platforms, their specifications will have
  solidified, and their implementations matured.
- The phases are as follows:
  1. Batch decryption system
  2. Naive rollup implementation
  3. Full rollup implementation

### Phase 1: Batch Decryption System

The batch decryption system (BDS) is a generic system that decrypts batches of
cipher transactions. It is oblivious to how the batches are generated or how
they will be processed.

The BDS consists of two types of nodes: Keypers and decryptors. Their identities
are registered in a registry contract on the main chain.

#### Keypers

Keypers take as input a stream of epoch end triggers and output a corresponding
stream of epoch decryption keys.

The keypers run a private Tendermint chain in order to achieve consenus among
them. On this chain, they generate the eon encryption key and one eon decryption
key for each keyper.

In addition, the keypers connect to two peer-to-peer gossip networks:

- the epoch end trigger network
- the epoch decryption key network

Whenever a keyper receives a trigger on the epoch end trigger network, they
authenticate it according to rules that are to be specified in later phases. If
the trigger is valid, they send their epoch decryption key share to the epoch
decryption key network. Keypers also pick up shares of their peers from the
network and, when they have enough of them, they broadcast the aggregate.

### Decryptors

Decryptors take as input a sequence of batches, each of which contains a list of
encrypted transactions, as well as the aggregated epoch decryption keys. They
output the decrypted batches as well as a signature authenticating the
decryption.

Decryptors receive their inputs from peer-to-peer gossip networks:

- Cipher batches from the cipher batch network (sourced by a TBD entity)
- Epoch decryption keys from the epoch decryption key network (sourced by the
  keypers)

All inputs are validated. If the decryptors receive a matching pair of cipher
batch and decryption key, they decrypt the batch and sign the result. Both
outputs are broadcast. Decryption signatures are also aggregated.

### Phase 2: Naive Rollup Implementation

Phase 2 integrates the BDS with a rollup in a practical way, accepting that a
limited amount of censorship is possible. It adds a collator node that creates
cipher transaction batches and hacks Optimism to only accept plaintext batches
in the right order and with a valid decryption signature.

The collator node provides a publicly accessible interface through which users
can submit enveloped encrypted transactions. They select transactions based on
their envelope and in regular intervals publish them as a batch. In addition,
they commit to the batch by signing its hash. Both batch and commitment are
broadcast on respective peer-to-peer networks.

The keyper is modified such that it listens for batch commitments from the
collator and accepts them as epoch end triggers. The decryptor is connected to
the batch broadcast network which it uses as input channel for cipher batches.

Optimism's executor contract is forked such that it executes only batches with
valid decryption signatures. In addition, the chain makes sure that the
execution context matches the batch commitment.

Note that phase 2 still allows for frontrunning via L1-to-L2 transactions.
Depending on the state of development of phase 3, there are multiple options to
deal with this:

- Ignore it, assuming phase 3 will be ready soon.
- Disable L1-to-L2 transactions. This will prevent frontrunning, but make the
  rollup unusable except for demo applications.
- Delay L1-to-L2 transactions via a special contract. This will prevent
  frontrunning, but worsen the user experience.

### Phase 3: Full Rollup Implementation

Phase 3 improves on the naive rollup implementation by preventing censorship by
individual privileged actors.

Phase 2 allows the collator to censor transactions (albeit only while they are
still encrypted). To solve this issue, the full implementation provides an
alternative path for plaintext transactions to be included in the rollup without
having to go through the collator. This path originates on the mainchain, but is
significantly delayed, such that it cannot be used to frontrun batch commitment.
