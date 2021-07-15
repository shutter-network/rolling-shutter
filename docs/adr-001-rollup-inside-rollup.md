# Rollup inside Rollup

- Author: Ralf Schmitt
- Status: work in progress
- Date: 2021-07-19

## Overview

This document describes a high-level view of a possible Shutter
implementation using nested arbitrum rollups. We describe some prior
ideas first, before presenting the final proposal.

## Prior Work

### The 2 week simple contracts only solution

The idea here was pretty simple. We run a rollup and deploy a set of
contracs to it. These contracts handle most of the work (besides the
distributed key generation). We would change the rollup implementation
to only allow transactions to any of the contracts.

The following contracts are needed:

#### Transaction Buffer Contract

`addTransaction(bytes memory transaction, uint64 epoch, bool isEncrypted)`
stores txs in state
ordered by epoch and submission time (how to deal with past epochs?)

#### Key Contract

`submitDecryptionKeyShare(bytes memory share, uint64 epoch)`
called by keypers to add decryption key share for particular epoch
`computeDecryptionKey(uint64 epoch)`
called once enough shares have been submitted and the key can be generated
`submitEonKey(bytes memory key, uint64 startEpoch, bytes[] memory votes)`
called for every eon to submit the eon key, with votes by keypers

#### Executor Contract

`executeTransactions(uint64 startIndex, uint64 maxNum)`
fetch encrypted tx from transaction buffer
fetch key from key contract
decrypt and execute tx (c.f. meta txs)
repeat, abort if key not available yet and not timed out

#### Keyper Set Contract

Manages keyper set -- how does it know what’s right?

### Why it doesn't work

After the decryption key for a certain epoch becomes available, users
can submit transactions to the rollup before `executeTransactions` has
been called. That opens a possible side channel because they can
e.g. change the balance of certain accounts and thereby influence the
execution of the decrypted transactions.

### The collator based solution

The idea here is to disallow transactions between the time when the
decryption key becomes available and when the decrypted transactions
are executed.

We run a collator that collects encrypted transactions from
users. Only the collator is allowed to submit transactions to the
arbitrum sequencer.

The collator commits to executing the decrypted transactions in a
certain order without adding additional transactions. When the keypers
see that commitment, they release the decryption key.
We slash the collator, if he doesn't fulfull his commitment.

#### Problems with this approach

Slashing the collator in this scenario is complicated, because we have
to look at the actual transactions he is executing on the rollup.

## Nested Rollups

The basic idea is to separate the collection of encrypted transactions
including the key handling from execution of decrypted transactions
into different environments.

### Collator chain

We deploy a set of contracts similar to the '2 weeks simple contract
solution' to an arbitrum rollup chain running in sequencer mode. We
call this chain the collator chain. We do not execute any decrypted
transactions on this chain. Instead of the executor contract, we
deploy a decryption contract. When the decryption key for a certain
epoch becomes available, the decryption contract decrypts all of the
encrypted transactions and submits them into an Inbox smart contract.
Only the decryption contract can submit transactions to the Inbox
contract.

Since the chain is running in sequencer mode, we can rely on instant
finality. Please note however, that arbitrum currently doesn't have a
way to enforce that, i.e. the sequencer could reorder transactions at
will. We will introduce a slashing mechanism: The sequencer will have
to commit to a certain sequence of transactions and will be slashed
when he doesn't fulfill it's commitment.

### Execution Chain

We deploy another rollup in non-sequencer mode onto the Collator
Chain. This Rollup uses the decryption contract's Inbox as it's
rollup Inbox.

## Possible Side Channels

The sequencer may influence the block number and timestamp on the
Collator Chain by delaying execution. We use the timestamp and
block number of the last encrypted transaction in each epoch instead
of the block number and timestamp when adding decrypted transactions
to the Inbox.

XXX What other information from them collator chain ends up in the
execution chain?
