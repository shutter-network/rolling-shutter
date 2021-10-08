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

-- name: GetDecryptionSignatures :many
SELECT * FROM decryptor.decryption_signature
WHERE epoch_id = $1 AND signed_hash = $2;

-- name: GetDecryptionSignature :one
SELECT * FROM decryptor.decryption_signature
WHERE epoch_id = $1 AND signers_bitfield = $2;

-- name: InsertDecryptionSignature :execresult
INSERT INTO decryptor.decryption_signature (
    epoch_id, signed_hash, signers_bitfield, signature
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
    identity.bls_public_key
FROM (
    SELECT
        start_epoch_id,
        index,
        address
    FROM decryptor.decryptor_set_member
    WHERE start_epoch_id = (
        SELECT
            m.start_epoch_id
        FROM decryptor.decryptor_set_member AS m
        WHERE m.start_epoch_id <= $1
        ORDER BY m.start_epoch_id DESC
        LIMIT 1
    )
) AS member
LEFT OUTER JOIN decryptor.decryptor_identity AS identity
ON member.address = identity.address
ORDER BY index;

-- name: GetDecryptorIndex :one
SELECT index
FROM decryptor.decryptor_set_member
WHERE start_epoch_id <= $1 AND address = $2;

-- name: GetDecryptorKey :one
SELECT bls_public_key FROM decryptor.decryptor_identity WHERE address = (
    SELECT address FROM decryptor.decryptor_set_member
    WHERE index = $1 AND start_epoch_id <= $2 ORDER BY start_epoch_id DESC LIMIT 1
);

-- name: InsertEonPublicKey :exec
INSERT INTO decryptor.eon_public_key (
    start_epoch_id,
    eon_public_key
) VALUES (
    $1, $2
);

-- name: GetEonPublicKey :one
SELECT eon_public_key
FROM decryptor.eon_public_key
WHERE start_epoch_id <= $1
ORDER BY start_epoch_id DESC
LIMIT 1;

-- name: InsertKeyperSet :exec
INSERT INTO decryptor.keyper_set (
    start_epoch_id,
    keypers,
    threshold
) VALUES (
    $1, $2, $3
);

-- name: InsertMeta :exec
INSERT INTO decryptor.meta_inf (key, value) VALUES ($1, $2);

-- name: GetMeta :one
SELECT * FROM decryptor.meta_inf WHERE key = $1;
