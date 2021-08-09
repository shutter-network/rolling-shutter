# Transactions in Rolling Shutter

- Author: Jannik Luhn
- Status: proposed
- Date: 2021-08-05

This document proposes transaction formats for Rolling Shutter.

## Context

The user wants to call a contract on the executor rollup. To do so, the
transaction has to be added to a batch in a certain epoch (in encrypted form) on
the sequencer rollup. Since most users won't have funds on the sequencer rollup,
they are unable to pay the required transaction fee and, thus, want to delegate
this task to relayers. In exchange, they pay relayers a fee on the sequencer
rollup. The relationship between users and relayers should be trustless, i.e.,

- relayers should be guaranteed to be paid the agreed upon fee if they submit
  the transaction
- users should be guaranteed to not pay the fee if the relayer does not submit
  the transaction

## Overview

The proposed transaction format consists of three nested layers, each with their
own transaction type, from inner to outer:

- the execution transaction
- the relay transaction
- the sequencer transaction

The execution transaction is signed by the user and processed by their account
contract on the execution rollup. It contains the contract call the user wants
to perform. Before passing it to the executor rollup, the sequencer rollup
performs some initial validation steps, too.

The relay transaction is signed by the user and is processed by the sequencer
contract on the sequencer rollup. It contains the encrypted execution
transaction and a fee payment to the eventual relayer.

Finally, the sequencer transaction is a standard Ethereum transaction signed by
the relayer that requests the sequencer contract on the sequencer rollup to
process the relay transaction.

In the following, we describe the format of each transaction and how they should
be processed.

## Sequencer Transaction

The sequencer transaction is a standard Ethereum transaction. It contains a
relay transaction as payload and calls the relayer contract with it.

## Relay Transaction

### Format

The signed relay transaction has the following format, understood to be ABI
encoded:

```c
struct RelayTransaction {
    epochIndex uint32;
    relayerFeeGWei uint64;
    encryptedTransaction bytes;
    v uint8;
    r bytes32;
    s bytes32;
}
```

Users create it by encrypting an `ExecutionTransaction` described below and
signing the following data according to
[EIP 712](https://eips.ethereum.org/EIPS/eip-712):

```c
struct UnsignedRelayTransaction {
    epochIndex uint32;
    relayerFeeGWei uint64;
    encryptedTransaction bytes;
}
```

The following domain separator is used:

```json
{
    name: "Rolling Shutter Relay Transaction",
    version: "0",
    chainId: <chain id of sequencer rollup>,
    verifyingContract: <address of relayer contract on sequencer rollup>,
    salt: "0598599f45ec70b3f60153f7e3249e02fe116cdb5d7df5c9570a677911dcac39"
}
```

### Processing

The relay transaction is handled by a relayer contract. It calls the batcher
contract and adds `encryptedTransaction` to the batch of epoch `epochIndex`.

If the transaction is successfully added, the contract registers a claim of
`relayerFeeGWei` to be paid from the signer's account to the caller of the
contract. The sequencer contract will transfer the claim to the execution
contract where the fee bank contract will settle it when the corresponding batch
is executed.

## Execution Transaction

### Format

The signed execution transaction has the following format, understood to be ABI
encoded:

```c
struct SignedExecutionTransaction {
    epochIndex uint32;
    nonce uint32;
    gasPriceGWei uint64;
    gasLimit uint32;
    to address;
    value uint256;
    data bytes;
    uint8 v;
    bytes32 r;
    bytes32 s;
}
```

Users create it by signing the following data structure according to EIP 712:

```c
struct UnsignedExecutionTransaction {
    epochIndex uint32;
    nonce uint32;
    gasPriceGWei uint64;
    gasLimit uint32;
    to address;
    value uint256;
    data bytes;
}
```

The following domain separator is used:

```json
{
    name: "Rolling Shutter Execution Rollup Transaction",
    version: "0",
    chainId: <chain id of sequencer rollup>,
    verifyingContract: <address of sequencer contract on sequencer rollup>,
    salt: "ad7348a8e8cc607f2cd4ae7059f15235a9ab5845f5981d0f7ebb260286e6f644"
}
```

### Processing

The execution transaction is processed twice: First, on the sequencer rollup in
the sequencer contract right after it has been decrypted, and a second time on
the execution rollup in an account contract.

The sequencer contract on the sequencer rollup first verifies that `epochIndex`
matches the currently executed batch. If it does, the contract sends the
transaction to the inbox of the execution rollup.

The account contract on the execution rollup executes the transaction as a
standard transaction, similar to Optimism's `OVM_ECDSAContractAccount` but
taking into account the different transaction format.

## Rationale

- EIP 712 was chosen as the signature scheme due to its support by end-user
  wallets, in particular Metamask.
- It was considered to add byte prefixes similar to
  [EIP 2718](https://eips.ethereum.org/EIPS/eip-2718), but rejected as it would
  interfere with EIP 712.
- Wherever possible, we tried to opt for concise encoding to keep transaction
  size as small as possible in order to save gas fees.
- Relayers can frontrun each other right now and steal transactions. This is
  deemed acceptable for now. If necessary, it can be prevented by allowing users
  to specify a relayer address (but this would degrade the user experience).
- The domain separator of the execution transaction uses chain id and contract
  address of the sequencer contract on the sequencer rollup. Alternatively, we
  could use the chain id of the execution rollup, but there we wouldn't have a
  canonical contract address.
