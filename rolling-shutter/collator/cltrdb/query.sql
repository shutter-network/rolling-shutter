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
INSERT INTO transaction (tx_id, epoch_id, encrypted_tx) VALUES ($1, $2, $3);

-- name: GetTransactionsByEpoch :many
SELECT encrypted_tx FROM transaction WHERE epoch_id = $1 ORDER BY tx_id;

-- name: SetNextEpochID :exec
INSERT INTO next_epoch (epoch_id) VALUES ($1)
ON CONFLICT (enforce_one_row) DO UPDATE
SET epoch_id = $1;

-- name: GetNextEpochID :one
SELECT epoch_id FROM next_epoch LIMIT 1;

-- name: InsertCandidateEonIfNotExists :exec
INSERT INTO eon (activation_block_number, eon_public_key, threshold) VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: InsertEonPublicKeyMessage :exec
INSERT INTO eon_public_key_message
    (eon_public_key, activation_block_number, keyper_index, msg_bytes)
    VALUES ($1, $2, $3, $4);

-- name: GetEonForBlock :one
SELECT * FROM eon
WHERE activation_block_number <= sqlc.arg(block_number)
-- There could be ambiguities when same activation_block_number
--   is used for different pubkeys, see #238
ORDER BY activation_block_number DESC
LIMIT 1;

-- name: GetEonPublicKeyMessages :many
WITH t3 AS (
	SELECT t2.activation_block_number, t2.eon_public_key, t2.msg_bytes, t2.keyper_index FROM (
		SELECT t1.num_signatures,
			t1.activation_block_number,
			t1.eon_public_key,
			t1.msg_bytes,
			t1.keyper_index
			FROM (
			SELECT eon.threshold,
				epkm.keyper_index,
				epkm.msg_bytes,
				epkm.activation_block_number,
				epkm.eon_public_key,
				COUNT(keyper_index) OVER (PARTITION BY (epkm.activation_block_number, epkm.eon_public_key)) num_signatures
			FROM eon_public_key_message epkm
			INNER JOIN eon
				ON epkm.activation_block_number = eon.activation_block_number
				AND epkm.eon_public_key = eon.eon_public_key
			WHERE epkm.activation_block_number <= sqlc.arg(block_number)
			) t1
		WHERE t1.num_signatures >= t1.threshold
		) t2
)
SELECT m.activation_block_number, m.eon_public_key, m.msg_bytes, m.keyper_index
FROM t3 m
WHERE NOT EXISTS (SELECT * FROM t3 b WHERE b.activation_block_number > m.activation_block_number)
;
