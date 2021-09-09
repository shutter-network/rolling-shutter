-- name: GetDecryptionKey :one
SELECT * FROM keyper.decryption_key
WHERE epoch_id = $1;
