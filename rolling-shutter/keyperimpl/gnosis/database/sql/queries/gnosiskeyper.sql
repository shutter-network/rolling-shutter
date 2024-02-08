-- name: InsertTransactionSubmittedEvent :execresult
INSERT INTO transaction_submitted_event (
    block_number,
    block_hash,
    tx_index,
    log_index,
    eon,
    identity_prefix,
    sender,
    gas_limit
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT DO NOTHING;

-- name: SetTransactionSubmittedEventsSyncedUntil :exec
INSERT INTO transaction_submitted_events_synced_until (block_number) VALUES ($1)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_number = $1;

-- name: GetTransactionSubmittedEventsSyncedUntil :one
SELECT block_number FROM transaction_submitted_events_synced_until LIMIT 1;

-- name: GetLocalTxPointer :one
SELECT local, local_block FROM tx_pointer LIMIT 1;

-- name: GetConsensusTxPointer :one
SELECT consensus, consensus_block FROM tx_pointer LIMIT 1;

-- name: UpdateLocalTxPointer :exec
INSERT INTO tx_pointer (local, local_block)
VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET local = $1, local_block = $2;

-- name: UpdateConsensusTxPointer :exec
INSERT INTO tx_pointer (consensus, consensus_block)
VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET consensus = $1, consensus_block = $2;
