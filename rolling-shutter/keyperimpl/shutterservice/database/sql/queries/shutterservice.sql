-- name: GetNotDecryptedIdentityRegisteredEvents :many
SELECT * FROM identity_registered_event
WHERE timestamp >= $1
ORDER BY index ASC;

-- name: GetIdentityRegisteredEventsSyncedUntil :one
SELECT * FROM identity_registered_events_synced_until LIMIT 1;
