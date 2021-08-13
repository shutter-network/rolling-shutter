# Execution Manager Enforced State Transition

- Author: Jannik Luhn
- Status: proposed
- Date: 2021-08-13

This document proposes another mechanism how Rolling Shutter might work.

## Overview

There's a single Optimism-based rollup that is responsible both for first
collecting encrypted transactions and for later decrypting and executing them.
The system operates on epochs, each of which has its own key pair and
transaction batch. Transactions can be added to batches until they are closed at
the end of their epoch. Closing is triggered by a dedicated collator.

Upon closing, the keypers will generate the decryption key off-chain. The key is
sent to the rollup and the transactions in the corresponding batch are decrypted
and executed. The process repeats for every epoch.

Frontrunning is prevented by blocking any state changing transaction to be
processed between closing and execution of a batch -- the only transactions
allowed in this timespan are those advancing the execution state. The system
ensures that these transactions are perfectly deterministic such that there is
no way of smuggling in information.

## Transaction Queuing

Encrypted transactions are stored in the state of the batcher contract,
organized by epoch. The collator can close a batch when its end time is
exceeded. Batches have to be closed in order.

Both transactions and batches have gas limits. A transaction can only be added
to a batch if the sum of gas limits of all transactions in the batch would not
exceed the batch gas limit.

When a transaction is added to a batch, `gas price * gas limit` is immediately
withdrawn from the transaction signer's account. Only the collator can add
transactions to a batch. This prevents attackers using up all gas with zero gas
price transactions.

## Transaction Execution

After a batch is closed, the transactions are decrypted and executed in the
order they have been added. This is triggered by repeated alternating calls to
the decryption and execution contracts. The execution manager ensures that no
other transactions are processed until execution is finished.

The decryption and execution is fully deterministic, in particular:

- transactions have no sender so that no fee is paid and no nonce is increased
- failing calls, e.g. because of a wrong decryption key, leave no trace
