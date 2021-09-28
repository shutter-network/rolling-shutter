-- name: GetDecryptionKey :one
SELECT * FROM keyper.decryption_key
WHERE epoch_id = $1;

-- name: InsertMeta :exec
INSERT INTO keyper.meta_inf (key, value) VALUES ($1, $2);

-- name: GetMeta :one
SELECT value FROM keyper.meta_inf WHERE key = $1;

-- name: InsertBatchConfig :exec
INSERT INTO keyper.tendermint_batch_config (config_index, height, keypers, threshold)
VALUES ($1, $2, $3, $4);

-- name: CountBatchConfigs :one
SELECT count(*) FROM keyper.tendermint_batch_config;

-- name: GetLatestBatchConfig :one
SELECT *
FROM keyper.tendermint_batch_config
ORDER BY config_index DESC
LIMIT 1;

-- name: GetBatchConfigs :many
SELECT *
FROM keyper.tendermint_batch_config
ORDER BY config_index;

-- name: GetBatchConfig :one
SELECT *
FROM keyper.tendermint_batch_config
WHERE config_index = $1;

-- name: TMSetSyncMeta :exec
INSERT INTO keyper.tendermint_sync_meta (current_block, last_committed_height, sync_timestamp)
VALUES ($1, $2, $3);

-- name: TMGetSyncMeta :one
SELECT *
FROM keyper.tendermint_sync_meta
ORDER BY current_block DESC, last_committed_height DESC
LIMIT 1;

-- name: InsertPureDKG :exec
INSERT INTO keyper.puredkg (eon,  puredkg) VALUES ($1, $2);

-- name: UpdatePureDKG :exec
UPDATE keyper.puredkg
SET puredkg=$2 WHERE eon=$1;

-- name: InsertEncryptionKey :exec
INSERT INTO keyper.tendermint_encryption_key (address, encryption_public_key) VALUES ($1, $2);

-- name: GetEncryptionKeys :many
SELECT * from keyper.tendermint_encryption_key;

-- name: ScheduleShutterMessage :one
INSERT INTO keyper.tendermint_outgoing_messages (msg)
VALUES ($1)
RETURNING id;

-- name: GetNextShutterMessage :one
SELECT * from keyper.tendermint_outgoing_messages
ORDER BY id
LIMIT 1;

-- name: DeleteShutterMessage :exec
DELETE FROM keyper.tendermint_outgoing_messages where id=$1;
