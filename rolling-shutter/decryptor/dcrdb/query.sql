-- name: GetCipherBatch :one
SELECT * FROM decryptor.cipher_batch
WHERE epoch_id = $1;

-- name: InsertCipherBatch :execresult
INSERT INTO decryptor.cipher_batch (
    epoch_id, data
) VALUES (
    $1, $2
)
ON CONFLICT DO NOTHING;

-- name: GetDecryptionKey :one
SELECT * FROM decryptor.decryption_key
WHERE epoch_id = $1;

-- name: InsertDecryptionKey :execresult
INSERT INTO decryptor.decryption_key (
    epoch_id, key
) VALUES (
    $1, $2
)
ON CONFLICT DO NOTHING;
