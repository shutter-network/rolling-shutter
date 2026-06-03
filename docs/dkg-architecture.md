# DKG Architecture

Shuttermint — the Tendermint sidechain — is gone. DKG coordination now runs entirely on Gnosis Chain via an Ethereum smart contract. This document explains the new code for someone familiar with the old architecture who wants to review the implementation.

## What was removed

- `rolling-shutter/app/` — the Tendermint ABCI application
- `rolling-shutter/shmsg/` — Shuttermint protobuf messages
- `rolling-shutter/keyper/shutterevents/` — Shuttermint event parsing
- `rolling-shutter/keyper/smobserver/` — Shuttermint state machine observer
- `rolling-shutter/keyper/fx/` — Shuttermint message sender
- `rolling-shutter/eonkeypublisher/` — EonKeyPublish contract interaction
- `rolling-shutter/cmd/chain/`, `cmd/bootstrap/` — chain and bootstrapping commands

## What was added

| Package | Role |
|---|---|
| `rolling-shutter/dkg/` | All DKG participation logic (dealing, accusing, apologizing, finalizing) |
| `rolling-shutter/txsender/` | Generic transaction outbox sender |
| `medley/chainsync/syncer/dkg.go` | Discovers DKG contracts from KeyperSetManager |
| `medley/chainsync/syncer/dkg_contract.go` | Subscribes to events on one DKG contract |
| `medley/chainsync/syncer/ecies.go` | Subscribes to ECIES Key Registry events |
| `rolling-shutter/contract/binding_dkg.abigen.gen.go` | Generated Go bindings for DKG and ECIES contracts (temporarily vendored here; will be imported from an external package like `github.com/shutter-network/shop-contracts/bindings` provides `KeyperSetManager`) |

The host keyper implementations (`keyperimpl/gnosis` and `keyperimpl/shutterservice`) wire these together; changes are structurally identical in both.

---

## Database schema changes

The migration files (`keyper/database/sql/migrations/`) tell the full story. Key changes:

**Dropped (Shuttermint remnants):**
- `tendermint_batch_config`, `tendermint_encryption_key`, `tendermint_outgoing_messages`, `tendermint_sync_meta`
- `poly_evals`, `puredkg` — old blob-based DKG storage
- `outgoing_eon_keys`, `last_batch_config_sent`, `last_block_seen`

**Added:**

| Table | Key columns | Purpose |
|---|---|---|
| `ecies_keys` | `keyper_address`, `ecies_public_key` | Keyper encryption keys, written by `ECIESKeySyncer` |
| `dkg_poly_commitments` | `(ksi, retry, keyper_index)` | Feldman commitments received from chain |
| `dkg_poly_evals` | `(ksi, retry, sender_index, receiver_index)` | Encrypted evaluations received from chain |
| `dkg_accusations` | `(ksi, retry, accuser_index, accused_index)` | Accusations received from chain |
| `dkg_apologies` | `(ksi, retry, apologizer_index, accuser_index)` | Apologies received from chain |
| `dkg_initial_states` | `(ksi, retry)`, `puredkg_bytes` | Serialized puredkg state after polynomial generation (see below) |
| `dkg_sent_actions` | `(ksi, retry, action)`, `tx_outbox_id` | Idempotency log for phase handler submissions |
| `tx_outbox` | `to_address`, `data`, `value`, `status`, `tx_hash` | Pending outbound transactions |

**Modified:**
- `eons` — extended with `dkg_contract` (address), `phase_length`, `lead_length`; `height` column dropped

`ksi` = `keyper_config_index` throughout.

---

## Data flow

### 1. Keyper Set registration

When `DKGSyncer` receives a `KeyperSetAdded` event, `keyperimpl/*/newkeyperset.go:processNewKeyperSet` runs:

1. Inserts the keyper set into `observer.keyper_set`.
2. If this keyper is a member, calls the DKG contract to read `PHASE_LENGTH` and `DKG_LEAD_LENGTH`, then inserts a row into `eons` with those parameters and the DKG contract address.
3. Calls `dkg.MaybeRegisterECIESKey` — if this keyper has no ECIES key registered on-chain, enqueues a `registerKey` transaction via `tx_outbox`.

The `eons` row is the sole source of phase parameters for the DKG module. It never calls the chain again.

### 2. Chain event ingestion

`DKGSyncer` (`medley/chainsync/syncer/dkg.go`) watches `KeyperSetManager` for new keyper sets. For each unique DKG contract address it discovers, it starts a `DKGContractSyncer` (`syncer/dkg_contract.go`) as a service. `DKGContractSyncer` subscribes to five event types on that contract and forwards them to the host keyper's handler:

| Event | Handler | Written to |
|---|---|---|
| `DealingSubmitted` | `newdkgevent.go` | `dkg_poly_commitments`, `dkg_poly_evals` |
| `AccusationSubmitted` | `newdkgevent.go` | `dkg_accusations` |
| `ApologySubmitted` | `newdkgevent.go` | `dkg_apologies` |
| `SuccessVoteSubmitted` | `newdkgevent.go` | `dkg_success_votes` |
| `DKGSucceeded` | `newdkgevent.go` | triggers `HandleDKGSuccess` |

`ECIESKeySyncer` (`syncer/ecies.go`) watches the ECIES Key Registry and writes registered keys to `ecies_keys`.

All event handlers write to the database and return. The DKG module never receives events directly.

### 3. Per-block participation

On every new block, the host keyper calls `dkg.Manager.HandleBlock(ctx, db, blockNumber)` (`dkg/manager.go`).

`HandleBlock` iterates every eon the keyper is a member of that has not yet succeeded:

1. Opens a **read-only transaction** to reconstruct the in-memory `puredkg` state by replaying all stored messages for this `(keyperConfigIndex, retryCounter)`.
2. Calls `PhaseAt(blockNumber)` (`dkg/phase.go`) to determine the current phase from pure arithmetic — no stored state.
3. Opens a **write transaction** and dispatches to the phase handler.

The two-transaction split is intentional: the read phase can be long (replaying all messages); the write transaction is kept as short as possible.

### 4. Phase handlers

Each handler in `dkg/dealing.go`, `dkg/accusing.go`, `dkg/apologizing.go`, `dkg/finalizing.go` follows the same pattern:

1. **Check idempotency** via `dkg_sent_actions`. If a row exists for `(keyperConfigIndex, retryCounter, action)`, return immediately — this block is a replay.
2. **Compute the action** using the in-memory `puredkg`.
3. **Enqueue to `tx_outbox`** (or skip if no action is needed, e.g. no accusations to make).
4. **Insert into `dkg_sent_actions`** with the outbox row id (or NULL if no tx was enqueued).

The `dkg_sent_actions` table is what prevents the same transaction from being submitted on every block while the phase is active.

**Dealing** generates the keyper's random polynomial and serialises the resulting `puredkg` state into `dkg_initial_states` before enqueuing. This is a cryptographic necessity: the random polynomial cannot be re-derived from the messages stored on-chain. Subsequent handlers load this snapshot and replay received messages on top of it rather than trying to reconstruct the polynomial from scratch.

### 5. Transaction submission

`TxSender` (`txsender/txsender.go`) runs as an independent service polling `tx_outbox` in two passes per tick:

**Submit pass** — rows with `status = 'pending'`:
- Fetches pending nonce, estimates gas, computes EIP-1559 fee cap.
- Marks the row `submitted` with the tx hash **before** broadcasting. This ensures a crash between mark and send leaves a recoverable row rather than a phantom.
- Calls `SendTransaction`.

**Confirm pass** — rows with `status = 'submitted'`:
- Calls `TransactionReceipt` with the stored hash.
- Marks `confirmed` on success, leaves unchanged if not yet mined.

`TxSender` is intentionally generic — it knows nothing about DKG. Retry logic, gas price bumps, and stuck-tx handling belong here, not in the DKG module.

### 6. DKG success

When `DKGContractSyncer` receives a `DKGSucceeded` event, `Manager.HandleDKGSuccess` runs:
- Inserts a `dkg_result` row marking the eon as succeeded.
- Rebuilds `puredkg` from stored messages at the Finalizing phase to compute the eon secret key share.
- `HandleBlock` skips this eon from the next block onward.

The `KeyBroadcastContract.broadcastEonKey` call is made by the DKG Contract itself when the success-vote threshold is reached — not by the keyper.

---

## Key invariants

**The DKG module never calls the chain.** All phase parameters, ECIES keys, and DKG messages arrive via the database. If a required row is missing (e.g. a peer's ECIES key), the handler skips rather than fetching it.

**`puredkg` is ephemeral.** It is rebuilt from database rows on every `HandleBlock` call (or loaded from the `dkg_initial_states` snapshot for post-Dealing phases). There is no long-lived in-memory DKG state to get out of sync.

**Phase windows are pure arithmetic.** `PhaseAt(block)` has no side effects and needs no stored state beyond the `eons` row. The full retry schedule is determined the moment `KeyperSetAdded` is processed.

**`dkg_sent_actions` is the idempotency boundary.** Any code path that enqueues a transaction must insert a `dkg_sent_actions` row in the same database transaction. Without this, a crash between submit and confirm would cause the handler to enqueue a duplicate on restart.

---

## Database tables

| Table | Written by | Read by |
|---|---|---|
| `eons` | `newkeyperset` handler | DKG module (phase params, contract addr) |
| `ecies_keys` | `ECIESKeySyncer` | DKG module (encrypt PolyEvals) |
| `dkg_poly_commitments` | `newdkgevent` handler | DKG module (rebuild puredkg) |
| `dkg_poly_evals` | `newdkgevent` handler | DKG module (rebuild puredkg) |
| `dkg_accusations` | `newdkgevent` handler | DKG module (rebuild puredkg) |
| `dkg_apologies` | `newdkgevent` handler | DKG module (rebuild puredkg) |
| `dkg_success_votes` | `newdkgevent` handler | DKG module (rebuild puredkg) |
| `dkg_initial_states` | Dealing handler | Accusing/Apologizing/Finalizing handlers |
| `dkg_sent_actions` | Phase handlers | Phase handlers (idempotency check) |
| `tx_outbox` | Phase handlers, `newkeyperset` | `TxSender` |
| `dkg_result` | `HandleDKGSuccess` | `HandleBlock` (skip check) |
