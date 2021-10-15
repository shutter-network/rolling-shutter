-- name: InsertDecryptionKey :exec
INSERT INTO keyper.decryption_key (epoch_id, decryption_key)
VALUES ($1, $2);

-- name: GetDecryptionKey :one
SELECT * FROM keyper.decryption_key
WHERE epoch_id = $1;

-- name: ExistsDecryptionKey :one
SELECT EXISTS (
    SELECT 1
    FROM keyper.decryption_key
    WHERE epoch_id = $1
);

-- name: InsertDecryptionKeyShare :exec
INSERT INTO keyper.decryption_key_share (epoch_id, keyper_index, decryption_key_share)
VALUES ($1, $2, $3);

-- name: SelectDecryptionKeyShares :many
SELECT * FROM keyper.decryption_key_share
WHERE epoch_id = $1;

-- name: GetDecryptionKeyShare :one
SELECT * FROM keyper.decryption_key_share
WHERE epoch_id = $1 AND keyper_index = $2;

-- name: ExistsDecryptionKeyShare :one
SELECT EXISTS (
    SELECT 1
    FROM keyper.decryption_key_share
    WHERE epoch_id = $1 AND keyper_index = $2
);

-- name: CountDecryptionKeyShares :one
SELECT count(*) FROM keyper.decryption_key_share
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

-- name: GetLastCommittedHeight :one
SELECT last_committed_height
FROM keyper.tendermint_sync_meta
ORDER BY current_block DESC, last_committed_height DESC
LIMIT 1;

-- name: InsertPureDKG :exec
INSERT INTO keyper.puredkg (eon, puredkg) VALUES ($1, $2)
ON CONFLICT (eon) DO UPDATE SET puredkg=EXCLUDED.puredkg;

-- name: SelectPureDKG :many
SELECT * FROM keyper.puredkg;

-- name: DeletePureDKG :exec
DELETE FROM keyper.puredkg WHERE eon=$1;


-- name: InsertEncryptionKey :exec
INSERT INTO keyper.tendermint_encryption_key (address, encryption_public_key) VALUES ($1, $2);

-- name: GetEncryptionKeys :many
SELECT * FROM keyper.tendermint_encryption_key;

-- name: ScheduleShutterMessage :one
INSERT INTO keyper.tendermint_outgoing_messages (description, msg)
VALUES ($1, $2)
RETURNING id;

-- name: GetNextShutterMessage :one
SELECT * from keyper.tendermint_outgoing_messages
ORDER BY id
LIMIT 1;

-- name: DeleteShutterMessage :exec
DELETE FROM keyper.tendermint_outgoing_messages WHERE id=$1;

-- name: InsertEon :exec
INSERT INTO keyper.eons (eon, height, batch_index, config_index)
VALUES ($1, $2, $3, $4);

-- name: GetEon :one
SELECT * FROM keyper.eons WHERE eon=$1;

-- name: GetEonForEpoch :one
SELECT * FROM keyper.eons
WHERE batch_index <= $1
ORDER BY batch_index DESC, height DESC
LIMIT 1;

-- name: InsertPolyEval :exec
INSERT INTO keyper.poly_evals (eon, receiver_address, eval)
VALUES ($1, $2, $3);

-- name: PolyEvalsWithEncryptionKeys :many
SELECT ev.eon, ev.receiver_address, ev.eval,
       k.encryption_public_key,
       eons.height
FROM keyper.poly_evals ev
INNER JOIN keyper.tendermint_encryption_key k
      ON ev.receiver_address = k.address
INNER JOIN keyper.eons eons
      ON ev.eon = eons.eon
ORDER BY ev.eon;

-- PolyEvalsWithEncryptionKeys could probably already delete the entries from the poly_evals table.
-- I wasn't able to make this work, because of bugs in sqlc

-- name: DeletePolyEval :exec
DELETE FROM keyper.poly_evals ev WHERE ev.eon=$1 AND ev.receiver_address=$2;

-- name: DeletePolyEvalByEon :execresult
DELETE FROM keyper.poly_evals ev WHERE ev.eon=$1;

-- name: InsertDKGResult :exec
INSERT INTO keyper.dkg_result (eon,success,error,pure_result)
VALUES ($1,$2,$3,$4);

-- name: GetDKGResult :one
SELECT * FROM keyper.dkg_result
WHERE eon = $1;
