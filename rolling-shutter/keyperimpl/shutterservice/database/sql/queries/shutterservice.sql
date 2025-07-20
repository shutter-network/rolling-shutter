-- name: GetNotDecryptedIdentityRegisteredEvents :many
SELECT * FROM identity_registered_event
WHERE timestamp >= $1 AND timestamp <= $2 AND decrypted = false
ORDER BY timestamp ASC;

-- name: GetIdentityRegisteredEventsSyncedUntil :one
SELECT * FROM identity_registered_events_synced_until LIMIT 1;

-- name: SetCurrentDecryptionTrigger :exec
INSERT INTO current_decryption_trigger (eon, triggered_block_number, identities_hash)
VALUES ($1, $2, $3)
ON CONFLICT (eon, triggered_block_number) DO UPDATE
SET triggered_block_number = $2, identities_hash = $3;

-- name: GetCurrentDecryptionTrigger :one
SELECT * FROM current_decryption_trigger
WHERE eon = $1 ORDER BY triggered_block_number DESC LIMIT 1;

-- name: InsertDecryptionSignature :exec
INSERT INTO decryption_signatures (eon, keyper_index, identities_hash, signature)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING;

-- name: GetDecryptionSignatures :many
SELECT * FROM decryption_signatures
WHERE eon = $1 AND identities_hash = $2
ORDER BY keyper_index ASC
LIMIT $3;

-- name: UpdateEventTriggerDecryptedFlag :exec
UPDATE event_trigger_registered_event
SET decrypted = TRUE
WHERE (eon, identity) IN (
    SELECT UNNEST($1::bigint[]), UNNEST($2::bytea[])
);

-- name: InsertEventTriggerRegisteredEvent :execresult
INSERT INTO event_trigger_registered_event (
    block_number,
    block_hash,
    tx_index,
    log_index,
    eon,
    identity_prefix,
    sender,
    definition,
    ttl,
    identity
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (identity_prefix, sender) DO UPDATE SET
block_number = $1,
block_hash = $2,
tx_index = $3,
log_index = $4,
sender = $7,
definition = $8,
ttl = $9,
identity = $10;


-- name: UpdateDecryptedFlag :exec
UPDATE identity_registered_event
SET decrypted = TRUE
WHERE (eon, identity) IN (
    SELECT UNNEST($1::bigint[]), UNNEST($2::bytea[])
);

-- name: InsertIdentityRegisteredEvent :execresult
INSERT INTO identity_registered_event (
    block_number,
    block_hash,
    tx_index,
    log_index,
    eon,
    identity_prefix,
    sender,
    timestamp,
    identity
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (identity_prefix, sender) DO UPDATE SET
block_number = $1,
block_hash = $2,
tx_index = $3,
log_index = $4,
sender = $7,
timestamp = $8,
identity = $9;

-- name: SetIdentityRegisteredEventSyncedUntil :exec
INSERT INTO identity_registered_events_synced_until (block_hash, block_number) VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_hash = $1, block_number = $2;

-- name: DeleteIdentityRegisteredEventsFromBlockNumber :exec
DELETE FROM identity_registered_event WHERE block_number >= $1;
