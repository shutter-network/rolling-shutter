-- name: GetDecryptionKey :one
SELECT *
FROM decryption_key
WHERE epoch_id = $1;

-- name: GetDecryptionKeyCount :one
SELECT COUNT(DISTINCT epoch_id)
FROM decryption_key;

-- name: InsertDecryptionKey :execrows
INSERT INTO decryption_key (
        epoch_id,
        key
) VALUES (
        $1, $2
)
ON CONFLICT DO NOTHING;

-- name: InsertEonPublicKey :execrows
INSERT INTO eon_public_key (
        eon_id,
        eon_public_key
) VALUES (
        $1, $2
)
ON CONFLICT DO NOTHING;


-- name: GetEonPublicKey :one
SELECT eon_public_key
FROM eon_public_key
WHERE eon_id = $1;


-- name: GetEonPublicKeyLatest :one
SELECT eon_id, eon_public_key
FROM eon_public_key
ORDER BY eon_id DESC
LIMIT 1;


-- name: GetEonCount :one
SELECT COUNT(DISTINCT eon_id)
FROM eon_public_key;
