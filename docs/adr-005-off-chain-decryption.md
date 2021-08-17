# Rolling Shutter with Off-Chain Decryption

- Author: Jannik Luhn
- Status: proposed
- Date: 2021-08-17

## Motivation

Shutter's decryption algorithm involves computing a pairing. While the EVM has
precompiles for checking equality of two or more pairings, it does not allow
computing just a single one and outputting the result. In theory, we could build
a Solidity implementation, but according to our estimates it would likely be too
gas-inefficient, even in a rollup (in particular, Optimism in which a single
transaction must not exceed the mainchain block gas limit). Future versions of
Ethereum will hopefully include a suitable precompile, but for now we have to
find an acceptable interim solution that does not rely on decrypting
transactions in the EVM.

This proposal therefore performs the decryption off-chain and only the plaintext
transactions being submitted to a rollup.

## Participants

- Users
- Collator
- Keypers
- Decryptors

## Transaction Lifecycle

1. A user creates a cipher transaction encrypted for a particular epoch.
2. The user sends the cipher transaction to the collator some time before the
   end of the epoch.
3. At the end of each epoch, the collator proposes and broadcasts a batch
   consisting of a selection of the transactions they received for this epoch.
4. Upon seeing the cipher batch, the keypers compute and publish the decryption
   key for the epoch.
5. The decryptors decrypt the cipher batch with the decryption key resulting in
   the plain batch and publish an aggregate or threshold signature for it.
6. The collator submits the plain batch and the decryption signature to the
   rollup which checks the signature and executes the batch.

Communication between users and the collator is direct. Collator, keypers, and
decryptors broadcast their messages on respective p2p gossip networks.

## Rollup Modifications

The rollup is modified to ensure batches are executed properly, in particular:

- The decryption signature is checked before a batch is executed.
- Batches are guaranteed to be executed in order and without skips.
- Batch execution can be split into multiple sub transactions.
- Failed attempts to advance batch execution do not leave a trace in the rollup
  state.

Transactions that are not part of a rollup are only executed with significant
delay, such that they cannot frontrun transactions from batches.

## Slashing Conditions

- The collator is slashed if they produce two cipher batches for the same epoch.
- Decryptors are slashed using a Kleros court decision if they produce a forged
  plain batch (i.e., a batch that does not correspond to the decrypted batch at
  the same epoch).

## Transaction Fees

Cipher transactions are wrapped in an envelope that contains a signature of the
user or a relayer as well as a fee amount. The envelope will be included in the
plain transactions as well and on execution a fee will be withdrawn from the
signer's account.

The system does not detect if the decryptors strip the envelope of the contained
transaction and wrap it around a different one. However, this is a slashable
offense and the Kleros court can reimburse affected users.

## Possible Misbehavior

- The keypers can frontrun if a threshold majority colludes.
- The collator can censor transactions, but only while they are still encrypted.
  This is undetectable, but users can use the bypass mechanism to force
  inclusion.
- The collator can refuse to submit a plain batch. In this case, anyone else,
  including affected users, can takeover and submit it.
- The collator can produce two competing batches and submit only one of them.
  However, this is a slashable offense.
- The decryptors can censor and reorder transactions in the plain batch as well
  as include their own if they collude. However, this is a slashable offense.
