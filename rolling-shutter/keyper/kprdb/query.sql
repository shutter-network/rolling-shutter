-- name: GetDecryptionKey :one
SELECT * FROM keyper.decryption_key
WHERE epoch_id = $1;

-- name: InsertMeta :exec
INSERT INTO keyper.meta_inf (key, value) VALUES ($1, $2);

-- name: GetMeta :one
SELECT * FROM keyper.meta_inf WHERE key = $1;
