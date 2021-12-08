-- name: InsertMeta :exec
INSERT INTO collator.meta_inf (key, value) VALUES ($1, $2);

-- name: GetMeta :one
SELECT * FROM collator.meta_inf WHERE key = $1;

-- name: InsertTrigger :exec
INSERT INTO collator.decryption_trigger (epoch_id, batch_hash) VALUES ($1, $2);

-- name: GetTrigger :one
SELECT * FROM collator.decryption_trigger WHERE epoch_id = $1;

-- name: GetLastBatchEpochID :one
SELECT epoch_id FROM collator.decryption_trigger ORDER BY epoch_id DESC LIMIT 1;

-- name: InsertTx :exec
INSERT INTO collator.transaction (tx_id, epoch_id, encrypted_tx) VALUES ($1, $2, $3);

-- name: GetTransactionsByEpoch :many
SELECT encrypted_tx FROM collator.transaction WHERE epoch_id = $1 ORDER BY tx_id;

-- name: InsertEpochID :exec
INSERT INTO collator.epoch_id (epoch_id) VALUES ($1);

-- name: GetBiggestEpochID :one
SELECT epoch_id FROM collator.epoch_id ORDER BY epoch_id LIMIT 1;
