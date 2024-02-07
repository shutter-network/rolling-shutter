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