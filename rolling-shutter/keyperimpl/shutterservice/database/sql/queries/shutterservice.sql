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
ON CONFLICT (eon, identity_prefix, sender) DO UPDATE SET
block_number = $1,
block_hash = $2,
tx_index = $3,
log_index = $4,
definition = $8,
ttl = $9,
identity = $10;


-- name: UpdateTimeBasedDecryptedFlags :exec
UPDATE identity_registered_event
SET decrypted = TRUE
WHERE (eon, identity) IN (
    SELECT UNNEST(@eons::bigint[]), UNNEST(@identities::bytea[])
);

-- name: UpdateEventBasedDecryptedFlags :exec
UPDATE event_trigger_registered_event
SET decrypted = TRUE
WHERE (eon, identity) IN (
    SELECT UNNEST(@eons::bigint[]), UNNEST(@identities::bytea[])
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

-- name: GetMultiEventSyncStatus :one
SELECT * FROM multi_event_sync_status LIMIT 1;

-- name: SetMultiEventSyncStatus :exec
INSERT INTO multi_event_sync_status (block_number, block_hash) VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_number = $1, block_hash = $2;

-- name: DeleteEventTriggerRegisteredEventsFromBlockNumber :exec
DELETE FROM event_trigger_registered_event WHERE block_number >= $1;

-- name: InsertFiredTrigger :exec
INSERT INTO fired_triggers (eon, identity_prefix, sender, block_number, block_hash, tx_index, log_index)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (eon, identity_prefix, sender) DO NOTHING;

-- name: DeleteFiredTriggersFromBlockNumber :exec
DELETE FROM fired_triggers WHERE block_number >= $1;

-- name: GetActiveEventTriggerRegisteredEvents :many
SELECT * FROM event_trigger_registered_event e
WHERE e.block_number + ttl >= @block_number -- TTL not expired at given block
AND e.decrypted = false  -- not decrypted yet
AND NOT EXISTS (  -- not fired yet
    SELECT 1 FROM fired_triggers t
    WHERE t.identity_prefix = e.identity_prefix
    AND t.sender = e.sender
);

-- name: GetUndecryptedFiredTriggers :many
SELECT
   f.identity_prefix,
   f.sender,
   f.block_number,
   f.block_hash,
   f.tx_index,
   f.log_index,
   e.eon AS eon,
   e.ttl AS ttl,
   e.identity AS identity,
   e.decrypted AS decrypted
FROM fired_triggers f
INNER JOIN event_trigger_registered_event e ON f.identity_prefix = e.identity_prefix AND f.sender = e.sender
WHERE NOT EXISTS (  -- not decrypted yet
    SELECT 1 FROM event_trigger_registered_event e
    WHERE e.identity_prefix = f.identity_prefix
    AND e.sender = f.sender
    AND e.decrypted = true
);
