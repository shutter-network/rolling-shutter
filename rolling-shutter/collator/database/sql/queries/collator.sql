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

-- name: GetEonPublicKey :one
SELECT * FROM eon_public_key_candidate
WHERE confirmed AND eon = $1
LIMIT 1;

-- name: FindEonPublicKeyForBlock :one
SELECT * FROM eon_public_key_candidate
WHERE confirmed AND activation_block_number <= sqlc.arg(blocknumber)
ORDER BY activation_block_number DESC, keyper_config_index DESC
LIMIT 1;

-- name: FindEonPublicKeyVotes :many
SELECT * FROM eon_public_key_vote WHERE hash=$1 ORDER BY sender;

-- name: InsertTrigger :exec
INSERT INTO decryption_trigger (epoch_id, batch_hash, l1_block_number) VALUES ($1, $2, $3);

-- name: UpdateDecryptionTriggerSent :exec
UPDATE decryption_trigger
SET sent=NOW()
WHERE epoch_id=$1;

-- name: GetUnsentTriggers :many
SELECT * FROM decryption_trigger
WHERE sent IS NULL
ORDER BY id ASC;

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
INSERT INTO transaction (tx_hash, epoch_id, tx_bytes, status) VALUES ($1, $2, $3, $4);

-- name: GetTransactionsByEpoch :many
SELECT * FROM transaction WHERE epoch_id = $1 ORDER BY id ASC;

-- name: GetNonRejectedTransactionsByEpoch :many
SELECT * FROM transaction WHERE status<>'rejected' AND epoch_id = $1 ORDER BY id ASC;

-- name: GetCommittedTransactionsByEpoch :many
SELECT * FROM transaction WHERE status = 'committed' AND epoch_id = $1 ORDER BY id ASC;

-- name: RejectNewTransactions :exec
UPDATE transaction
SET status='rejected'
WHERE epoch_id=$1 AND status='new';

-- name: SetTransactionStatus :exec
UPDATE transaction
SET status=$2
WHERE tx_hash = $1;

-- name: SetNextBatch :exec
INSERT INTO next_batch (epoch_id, l1_block_number) VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET epoch_id = $1, l1_block_number = $2;

-- name: GetNextBatch :one
SELECT * FROM next_batch LIMIT 1;

-- name: InsertBatchTx :exec
INSERT INTO batchtx (epoch_id, marshaled) VALUES ($1, $2);

-- name: GetUnsubmittedBatchTx :one
SELECT * FROM batchtx WHERE submitted=false;

-- name: SetBatchSubmitted :exec
UPDATE batchtx SET submitted=true WHERE submitted=false;
