-- name: InsertTransactionSubmittedEvent :execresult
INSERT INTO transaction_submitted_event (
    index,
    block_number,
    block_hash,
    tx_index,
    log_index,
    eon,
    identity_prefix,
    sender,
    gas_limit
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT DO NOTHING;

-- name: GetTransactionSubmittedEvents :many
SELECT * FROM transaction_submitted_event
WHERE eon = $1 AND index >= $2
ORDER BY index ASC
LIMIT $3;

-- name: SetTransactionSubmittedEventsSyncedUntil :exec
INSERT INTO transaction_submitted_events_synced_until (block_number) VALUES ($1)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_number = $1;

-- name: GetTransactionSubmittedEventsSyncedUntil :one
SELECT block_number FROM transaction_submitted_events_synced_until LIMIT 1;

-- name: SetTransactionSubmittedEventCount :exec
INSERT INTO transaction_submitted_event_count (eon, event_count)
VALUES ($1, $2)
ON CONFLICT (eon) DO UPDATE
SET event_count = $2;

-- name: GetTransactionSubmittedEventCount :one
SELECT event_count FROM transaction_submitted_event_count
WHERE eon = $1
LIMIT 1;

-- name: GetTxPointer :one
SELECT * FROM tx_pointer
WHERE eon = $1;

-- name: InitTxPointer :exec
INSERT INTO tx_pointer (eon, block, value)
VALUES ($1, $2, 0)
ON CONFLICT DO NOTHING;

-- name: SetTxPointer :exec
INSERT INTO tx_pointer (eon, block, value)
VALUES ($1, $2, $3)
ON CONFLICT (eon) DO UPDATE
SET block = $2, value = $3;

-- name: SetCurrentDecryptionTrigger :exec
INSERT INTO current_decryption_trigger (eon, block, tx_pointer, identities_hash)
VALUES ($1, $2, $3, $4)
ON CONFLICT (eon) DO UPDATE
SET block = $2, tx_pointer = $3, identities_hash = $4;

-- name: GetCurrentDecryptionTrigger :one
SELECT * FROM current_decryption_trigger
WHERE eon = $1;

-- name: InsertSlotDecryptionSignature :exec
INSERT INTO slot_decryption_signatures (eon, block, keyper_index, tx_pointer, identities_hash, signature)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetSlotDecryptionSignatures :many
SELECT * FROM slot_decryption_signatures
WHERE eon = $1 AND block = $2 AND tx_pointer = $3 AND identities_hash = $4
ORDER BY keyper_index ASC
LIMIT $5;
