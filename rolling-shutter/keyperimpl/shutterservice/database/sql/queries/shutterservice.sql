-- name: GetNotDecryptedIdentityRegisteredEvents :many
SELECT * FROM identity_registered_event
WHERE timestamp >= $1
ORDER BY index ASC;

-- name: GetIdentityRegisteredEventsSyncedUntil :one
SELECT * FROM identity_registered_events_synced_until LIMIT 1;

-- name: SetCurrentDecryptionTrigger :exec
INSERT INTO current_decryption_trigger (eon, last_block_number, identities_hash)
VALUES ($1, $2, $3)
ON CONFLICT (eon, last_block_number) DO UPDATE
SET last_block_number = $2, identities_hash = $3;

-- name: GetCurrentDecryptionTrigger :one
SELECT * FROM current_decryption_trigger
WHERE eon = $1;
