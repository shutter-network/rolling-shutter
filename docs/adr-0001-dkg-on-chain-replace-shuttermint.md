# ADR 0008: Replace Shuttermint with an Ethereum DKG Contract

## Status

Accepted

## Context

Keypers previously coordinated Distributed Key Generation via Shuttermint — a dedicated Tendermint sidechain. Running Shuttermint required operating a separate process alongside the keyper, executing a bootstrapping ceremony to initialise the chain, maintaining a P2P network between keyper nodes, and continuously synchronising state between Shuttermint and the management contracts on Gnosis Chain. The management contracts already lived on Gnosis Chain; DKG coordination could move there too.

## Decision

Replace Shuttermint entirely with a single Ethereum smart contract (the **DKG Contract**) on Gnosis Chain. The contract acts as a bulletin board: Keypers submit DKG messages as transactions, the contract emits them as events, and Keypers read back what their peers submitted. The Shuttermint binary, validator key management, Tendermint P2P network, bootstrapping ceremony, and all synchronisation code between Shuttermint and the management contracts are removed.

## Design decisions

**No on-chain cryptographic verification.** The DKG Contract enforces only that the sender is a Keyper Set member and that the message type matches the current phase. It performs no BLS signature checks or polynomial commitment verification. The cryptographic primitives required (BLS12-381 pairings) are not available as EVM precompiles, and gas costs would be prohibitive on any plausible execution model. Keypers detect misbehaviour — invalid or missing evaluations — off-chain and exclude bad actors from their local DKG computation.

**Stateless, block-number-derived phase windows.** Phase boundaries are computed from the DKG Contract parameters (`PHASE_LENGTH`, `DKG_LEAD_LENGTH`) and the Keyper Set's activation block: `activation_block - dkg_lead_length + r * cycle_length`. No start block is recorded when a DKG Instance begins. A Keyper that receives a `KeyperSetAdded` event has everything it needs to compute the full phase schedule without any further on-chain or off-chain state, which eliminates a whole class of sync problems.

**Per-Keyper-Set DKG Contract address.** The DKG Contract hardcodes `PHASE_LENGTH` and `DKG_LEAD_LENGTH`. When those parameters need to change, a new contract is deployed and the new Keyper Set points to it; when they stay the same, multiple Keyper Sets share one contract. The address is stored in `KeyperSet.sol` (`dkgContract` field) and read by the keyper when it processes a `KeyperSetAdded` event. This follows the same pattern as the `publisher` field already in `KeyperSet.sol`.

**DKG module as a DB-driven reactor.** The `rolling-shutter/dkg` package contains all DKG participation logic but owns no chain subscriptions. On each new block the host keyper calls `HandleBlock`; the module reads current state from the database and writes back any required actions (message rows, transaction outbox entries). The host keyper is responsible for delivering chain events into the database. This reuses the existing chainsync infrastructure rather than introducing a second subscription stack.

**Transaction outbox for decoupled sending.** DKG messages are not submitted to the chain inside open database transactions. Instead, the DKG module writes a row to `tx_outbox` as part of the same transaction as the associated message rows; a separate `TxSender` service reads pending rows and handles signing, nonce management, gas estimation, and resubmission. This separates DKG participation logic from transaction lifecycle concerns (retries, gas price bumps, stuck-tx handling), keeps database transactions short — inline RPC calls hold locks for an unbounded duration and create goroutine contention — and allows the outbox to serve ECIES key registration and success votes as well as DKG messages.

**ECIES keys persist across Keyper Set transitions.** Keypers register their ECIES encryption public key once in a standalone ECIES Key Registry contract; they do not re-register when joining a new Keyper Set. ECIES keys are long-lived identity keys that Keypers are not expected to rotate.

**DKG events delivered via chainsync subscriptions.** DKG Contract events and ECIES Registry events use the same `eth_subscribe` pattern already in place for `KeyperSetSyncer`, rather than block-range polling with reorg handling. Events emitted while a keyper is offline are not replayed on reconnect. This is consistent with the existing event syncing mechanism; block-range polling can be added later without changing the DKG module.

## Consequences

**Positive:**

- Eliminates the Shuttermint binary, validator key management, Tendermint P2P network, bootstrapping ceremony, and all synchronisation code between Shuttermint and the management contracts.
- All DKG coordination is observable on the same chain as the rest of the protocol.
- Phase windows are fully auditable from on-chain parameters alone.

**Negative:**

- DKG messages are Ethereum transactions, incurring gas costs per message. The number of messages scales quadratically with Keyper Set size (each Keyper sends one PolyEval per peer), which limits the maximum viable Keyper Set size more than the per-message cost alone would suggest.
- A keyper that is offline during a DKG phase will miss events and be unable to participate in that instance. It must wait for the next retry.
