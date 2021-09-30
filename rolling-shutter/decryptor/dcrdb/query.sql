-- name: GetCipherBatch :one
SELECT * FROM decryptor.cipher_batch
WHERE epoch_id = $1;

-- name: InsertCipherBatch :execresult
INSERT INTO decryptor.cipher_batch (
    epoch_id, transactions
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

-- name: InsertDecryptorIdentity :exec
INSERT INTO decryptor.decryptor_identity (
    address, bls_public_key
) VALUES (
    $1, $2
);

-- name: InsertDecryptorSetMember :exec
INSERT INTO decryptor.decryptor_set_member (
    start_epoch_id, index, address
) VALUES (
    $1, $2, $3
);

-- name: GetDecryptorSet :many
SELECT
    member.start_epoch_id,
    member.index,
    member.address,
    decryptor_identity.bls_public_key
FROM (
    SELECT * FROM decryptor.decryptor_set_member
    WHERE start_epoch_id <= $1
    ORDER BY start_epoch_id DESC
    FETCH FIRST ROW WITH TIES
) AS member
INNER JOIN decryptor.decryptor_identity
ON member.address=decryptor.decryptor_identity.address
ORDER BY index;

-- name: GetDecryptorIndex :one
SELECT index
FROM decryptor.decryptor_set_member
WHERE start_epoch_id <= $1 AND address = $2;

-- name: InsertMeta :exec
INSERT INTO decryptor.meta_inf (key, value) VALUES ($1, $2);

-- name: GetMeta :one
SELECT * FROM decryptor.meta_inf WHERE key = $1;
