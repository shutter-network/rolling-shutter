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

-- name: GetDecryptionSignature :one
SELECT * FROM decryptor.decryption_signature
WHERE epoch_id = $1 AND signer_index = $2;

-- name: InsertDecryptionSignature :execresult
INSERT INTO decryptor.decryption_signature (
    epoch_id, signed_hash, signer_index, signature
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT DO NOTHING;

-- name: InsertMeta :exec
INSERT INTO decryptor.meta_inf (key, value) VALUES ($1, $2);

-- name: GetMeta :one
SELECT * FROM decryptor.meta_inf WHERE key = $1;
