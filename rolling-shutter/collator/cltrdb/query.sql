-- name: InsertTrigger :exec
INSERT INTO decryption_trigger (epoch_id, batch_hash) VALUES ($1, $2);

-- name: GetTrigger :one
SELECT * FROM decryption_trigger WHERE epoch_id = $1;

-- name: GetLastBatchEpochID :one
SELECT epoch_id FROM decryption_trigger ORDER BY epoch_id DESC LIMIT 1;

-- name: InsertTx :exec
INSERT INTO transaction (tx_id, epoch_id, encrypted_tx) VALUES ($1, $2, $3);

-- name: GetTransactionsByEpoch :many
SELECT encrypted_tx FROM transaction WHERE epoch_id = $1 ORDER BY tx_id;

-- name: SetNextEpochID :exec
INSERT INTO next_epoch (epoch_id, block_number) VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET epoch_id = $1, block_number = $2;

-- name: GetNextEpochID :one
SELECT epoch_id, block_number FROM next_epoch LIMIT 1;
