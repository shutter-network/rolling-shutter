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
ON CONFLICT (index, eon) DO UPDATE SET
block_number = $2,
block_hash = $3,
tx_index = $4,
log_index = $5,
identity_prefix = $7,
sender = $8,
gas_limit = $9;

-- name: GetTransactionSubmittedEvents :many
SELECT * FROM transaction_submitted_event
WHERE eon = $1 AND index >= $2 AND index < $2 + $3
ORDER BY index ASC
LIMIT $3;

-- name: SetTransactionSubmittedEventsSyncedUntil :exec
INSERT INTO transaction_submitted_events_synced_until (block_hash, block_number, slot) VALUES ($1, $2, $3)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_hash = $1, block_number = $2, slot = $3;

-- name: GetTransactionSubmittedEventsSyncedUntil :one
SELECT * FROM transaction_submitted_events_synced_until LIMIT 1;

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
INSERT INTO tx_pointer (eon, age, value)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: SetTxPointer :exec
INSERT INTO tx_pointer (eon, age, value)
VALUES ($1, $2, $3)
ON CONFLICT (eon) DO UPDATE
SET age = $2, value = $3;

-- name: IncrementTxPointerAge :one
UPDATE tx_pointer
SET age = age + 1
WHERE eon = $1
RETURNING age;

-- name: ResetAllTxPointerAges :exec
UPDATE tx_pointer
SET age = NULL;

-- name: SetCurrentDecryptionTrigger :exec
INSERT INTO current_decryption_trigger (eon, slot, tx_pointer, identities_hash)
VALUES ($1, $2, $3, $4)
ON CONFLICT (eon) DO UPDATE
SET slot = $2, tx_pointer = $3, identities_hash = $4;

-- name: GetCurrentDecryptionTrigger :one
SELECT * FROM current_decryption_trigger
WHERE eon = $1;

-- name: InsertSlotDecryptionSignature :exec
INSERT INTO slot_decryption_signatures (eon, slot, keyper_index, tx_pointer, identities_hash, signature)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING;

-- name: GetSlotDecryptionSignatures :many
SELECT * FROM slot_decryption_signatures
WHERE eon = $1 AND slot = $2 AND tx_pointer = $3 AND identities_hash = $4
ORDER BY keyper_index ASC
LIMIT $5;

-- name: InsertValidatorRegistration :exec
INSERT INTO validator_registrations (
    block_number,
    block_hash,
    tx_index,
    log_index,
    validator_index,
    nonce,
    is_registration
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: IsValidatorRegistered :one
SELECT is_registration FROM validator_registrations
WHERE validator_index = $1 AND block_number < $2
ORDER BY block_number DESC, tx_index DESC, log_index DESC
LIMIT 1;

-- name: SetValidatorRegistrationsSyncedUntil :exec
INSERT INTO validator_registrations_synced_until (block_hash, block_number) VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_hash = $1, block_number = $2;

-- name: GetValidatorRegistrationsSyncedUntil :one
SELECT * FROM validator_registrations_synced_until LIMIT 1;

-- name: GetValidatorRegistrationNonceBefore :one
SELECT nonce FROM validator_registrations
WHERE validator_index = $1 AND block_number <= $2 AND tx_index <= $3 AND log_index <= $4
ORDER BY block_number DESC, tx_index DESC, log_index DESC
LIMIT 1;

-- name: GetNumValidatorRegistrations :one
SELECT COUNT(*) FROM validator_registrations;