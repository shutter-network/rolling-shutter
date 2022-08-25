-- name: InsertDecryptionKey :execresult
INSERT INTO decryption_key (eon, epoch_id, decryption_key)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetDecryptionKey :one
SELECT * FROM decryption_key
WHERE eon = $1 AND epoch_id = $2;

-- name: ExistsDecryptionKey :one
SELECT EXISTS (
    SELECT 1
    FROM decryption_key
    WHERE eon = $1 AND epoch_id = $2
);

-- name: InsertDecryptionKeyShare :exec
INSERT INTO decryption_key_share (eon, epoch_id, keyper_index, decryption_key_share)
VALUES ($1, $2, $3, $4);

-- name: SelectDecryptionKeyShares :many
SELECT * FROM decryption_key_share
WHERE eon = $1 AND epoch_id = $2;

-- name: GetDecryptionKeyShare :one
SELECT * FROM decryption_key_share
WHERE eon = $1 AND epoch_id = $2 AND keyper_index = $3;

-- name: ExistsDecryptionKeyShare :one
SELECT EXISTS (
    SELECT 1
    FROM decryption_key_share
    WHERE eon = $1 AND epoch_id = $2 AND keyper_index = $3
);

-- name: CountDecryptionKeyShares :one
SELECT count(*) FROM decryption_key_share
WHERE eon = $1 AND epoch_id = $2;

-- name: InsertBatchConfig :exec
INSERT INTO tendermint_batch_config (config_index, height, keypers, threshold, started, activation_block_number)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: CountBatchConfigs :one
SELECT count(*) FROM tendermint_batch_config;

-- name: GetLatestBatchConfig :one
SELECT *
FROM tendermint_batch_config
ORDER BY config_index DESC
LIMIT 1;

-- name: CountBatchConfigsInBlockRange :one
SELECT COUNT(*)
FROM tendermint_batch_config
WHERE @start_block <= activation_block_number AND activation_block_number < @end_block;

-- name: GetBatchConfigs :many
SELECT *
FROM tendermint_batch_config
ORDER BY config_index;

-- name: GetBatchConfig :one
SELECT *
FROM tendermint_batch_config
WHERE config_index = $1;

-- name: SetBatchConfigStarted :exec
UPDATE tendermint_batch_config SET started = TRUE
WHERE config_index = $1;

-- name: TMSetSyncMeta :exec
INSERT INTO tendermint_sync_meta (current_block, last_committed_height, sync_timestamp)
VALUES ($1, $2, $3);

-- name: TMGetSyncMeta :one
SELECT *
FROM tendermint_sync_meta
ORDER BY current_block DESC, last_committed_height DESC
LIMIT 1;

-- name: GetLastCommittedHeight :one
SELECT last_committed_height
FROM tendermint_sync_meta
ORDER BY current_block DESC, last_committed_height DESC
LIMIT 1;

-- name: InsertPureDKG :exec
INSERT INTO puredkg (eon, puredkg) VALUES ($1, $2)
ON CONFLICT (eon) DO UPDATE SET puredkg=EXCLUDED.puredkg;

-- name: SelectPureDKG :many
SELECT * FROM puredkg;

-- name: DeletePureDKG :exec
DELETE FROM puredkg WHERE eon=$1;

-- name: InsertEncryptionKey :exec
INSERT INTO tendermint_encryption_key (address, encryption_public_key) VALUES ($1, $2);

-- name: GetEncryptionKeys :many
SELECT * FROM tendermint_encryption_key;

-- name: ScheduleShutterMessage :one
INSERT INTO tendermint_outgoing_messages (description, msg)
VALUES ($1, $2)
RETURNING id;

-- name: GetNextShutterMessage :one
SELECT * from tendermint_outgoing_messages
ORDER BY id
LIMIT 1;

-- name: DeleteShutterMessage :exec
DELETE FROM tendermint_outgoing_messages WHERE id=$1;

-- name: InsertEon :exec
INSERT INTO eons (eon, height, activation_block_number, config_index)
VALUES ($1, $2, $3, $4);

-- name: GetEon :one
SELECT * FROM eons WHERE eon=$1;

-- name: GetEonForBlockNumber :one
SELECT * FROM eons
WHERE activation_block_number <= sqlc.arg(block_number)
ORDER BY activation_block_number DESC, height DESC
LIMIT 1;

-- name: GetAllEons :many
SELECT * FROM eons ORDER BY eon;

-- name: InsertPolyEval :exec
INSERT INTO poly_evals (eon, receiver_address, eval)
VALUES ($1, $2, $3);

-- name: PolyEvalsWithEncryptionKeys :many
SELECT ev.eon, ev.receiver_address, ev.eval,
       k.encryption_public_key,
       eons.height
FROM poly_evals ev
INNER JOIN tendermint_encryption_key k
      ON ev.receiver_address = k.address
INNER JOIN eons eons
      ON ev.eon = eons.eon
ORDER BY ev.eon;

-- PolyEvalsWithEncryptionKeys could probably already delete the entries from the poly_evals table.
-- I wasn't able to make this work, because of bugs in sqlc

-- name: DeletePolyEval :exec
DELETE FROM poly_evals ev WHERE ev.eon=$1 AND ev.receiver_address=$2;

-- name: DeletePolyEvalByEon :execresult
DELETE FROM poly_evals ev WHERE ev.eon=$1;

-- name: InsertDKGResult :exec
INSERT INTO dkg_result (eon,success,error,pure_result)
VALUES ($1,$2,$3,$4);

-- name: GetDKGResult :one
SELECT * FROM dkg_result
WHERE eon = $1;

-- name: GetDKGResultForBlockNumber :one
SELECT * FROM dkg_result
WHERE eon = (SELECT eon FROM eons WHERE activation_block_number <= sqlc.arg(block_number)
ORDER BY activation_block_number DESC, height DESC
LIMIT 1);

-- name: InsertEonPublicKey :exec
INSERT INTO outgoing_eon_keys (eon_public_key, eon)
VALUES ($1, $2);

-- name: GetAndDeleteEonPublicKeys :many
DELETE FROM outgoing_eon_keys RETURNING *;

-- name: SetLastBatchConfigSent :exec
INSERT INTO last_batch_config_sent (event_index) VALUES ($1)
ON CONFLICT (enforce_one_row) DO UPDATE
SET event_index = $1;

-- name: GetLastBatchConfigSent :one
SELECT event_index FROM last_batch_config_sent LIMIT 1;


-- name: SetLastBlockSeen :exec
INSERT INTO last_block_seen (block_number) VALUES ($1)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_number = $1;

-- name: GetLastBlockSeen :one
SELECT block_number FROM last_block_seen LIMIT 1;
