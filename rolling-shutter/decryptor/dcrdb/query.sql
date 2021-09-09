-- name: GetCipherBatch :one
SELECT * FROM decryptor.cipher_batch
WHERE epoch_id = $1;
