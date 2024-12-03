-- name: GetNotDecryptedIdentityRegisteredEvents :many
SELECT * FROM identity_registered_event
WHERE timestamp >= $1 AND timestamp <= $2
ORDER BY index ASC;

-- name: GetIdentityRegisteredEventsSyncedUntil :one
SELECT * FROM identity_registered_events_synced_until LIMIT 1;

-- name: SetCurrentDecryptionTrigger :exec
INSERT INTO current_decryption_trigger (eon, triggered_block_number, identities_hash)
VALUES ($1, $2, $3)
ON CONFLICT (eon, triggered_block_number) DO UPDATE
SET triggered_block_number = $2, identities_hash = $3;

-- name: GetCurrentDecryptionTrigger :one
SELECT * FROM current_decryption_trigger
WHERE eon = $1;

-- name: InsertDecryptionSignature :exec
INSERT INTO decryption_signatures (eon, keyper_index, identities_hash, signature)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING;

-- name: GetDecryptionSignatures :many
SELECT * FROM decryption_signatures
WHERE eon = $1 AND identities_hash = $2
ORDER BY keyper_index ASC
LIMIT $3;

-- name: UpdateDecryptedFlag :exec
UPDATE identity_registered_event
SET decrypted = TRUE
WHERE (eon, identity_prefix) IN (
    SELECT UNNEST($1::bigint[]), UNNEST($2::bytea[])
);