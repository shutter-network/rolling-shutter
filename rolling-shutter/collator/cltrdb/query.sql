-- name: InsertEonPublicKeyCandidate :exec
INSERT INTO eon_public_key_candidate
       (hash, eon_public_key, activation_block_number, keyper_config_index, eon)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING;

-- name: InsertEonPublicKeyVote :exec
INSERT INTO eon_public_key_vote
       (hash, sender, signature, eon, keyper_config_index)
VALUES ($1, $2, $3, $4, $5);

-- name: CountEonPublicKeyVotes :one
SELECT COUNT(*) from eon_public_key_vote WHERE hash=$1;

-- name: ConfirmEonPublicKey :exec
UPDATE eon_public_key_candidate
SET confirmed=TRUE
WHERE hash=$1;

-- name: FindEonPublicKeyForBlock :one
SELECT * FROM eon_public_key_candidate
WHERE confirmed AND activation_block_number <= sqlc.arg(blocknumber)
ORDER BY activation_block_number DESC, keyper_config_index DESC
LIMIT 1;

-- name: FindEonPublicKeyVotes :many
SELECT * FROM eon_public_key_vote WHERE hash=$1 ORDER BY sender;

-- name: InsertTrigger :exec
INSERT INTO decryption_trigger (epoch_id, batch_hash) VALUES ($1, $2);

-- name: GetTrigger :one
SELECT * FROM decryption_trigger WHERE epoch_id = $1;

-- name: InsertDecryptionKey :execresult
INSERT INTO decryption_key (epoch_id, decryption_key)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetDecryptionKey :one
SELECT * FROM decryption_key
WHERE epoch_id = $1;

-- name: ExistsDecryptionKey :one
SELECT EXISTS (
    SELECT 1
    FROM decryption_key
    WHERE epoch_id = $1
);

-- name: GetLastBatchEpochID :one
SELECT epoch_id FROM decryption_trigger ORDER BY epoch_id DESC LIMIT 1;

-- name: InsertTx :exec
INSERT INTO transaction (tx_hash, epoch_id, tx_bytes) VALUES ($1, $2, $3);

-- name: GetTransactionsByEpoch :many
SELECT tx_bytes FROM transaction WHERE epoch_id = $1 ORDER BY id ASC;

-- name: SetNextEpochID :exec
INSERT INTO next_epoch (epoch_id) VALUES ($1)
ON CONFLICT (enforce_one_row) DO UPDATE
SET epoch_id = $1;

-- name: GetNextEpochID :one
SELECT epoch_id FROM next_epoch LIMIT 1;
