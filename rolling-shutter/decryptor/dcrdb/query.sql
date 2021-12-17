-- name: GetCipherBatch :one
SELECT * FROM cipher_batch
WHERE epoch_id = $1;

-- name: InsertCipherBatch :execresult
INSERT INTO cipher_batch (
    epoch_id, transactions
) VALUES (
    $1, $2
)
ON CONFLICT DO NOTHING;

-- name: GetDecryptionKey :one
SELECT * FROM decryption_key
WHERE epoch_id = $1;

-- name: InsertDecryptionKey :execresult
INSERT INTO decryption_key (
    epoch_id, key
) VALUES (
    $1, $2
)
ON CONFLICT DO NOTHING;

-- name: GetDecryptionSignatures :many
SELECT * FROM decryption_signature
WHERE epoch_id = $1 AND signed_hash = $2;

-- name: GetDecryptionSignature :one
SELECT * FROM decryption_signature
WHERE epoch_id = $1 AND signers_bitfield = $2;

-- name: InsertDecryptionSignature :execresult
INSERT INTO decryption_signature (
    epoch_id, signed_hash, signers_bitfield, signature
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT DO NOTHING;

-- name: GetAggregatedSignature :one
SELECT * FROM aggregated_signature
WHERE signed_hash = $1;

-- name: ExistsAggregatedSignature :one
SELECT EXISTS(SELECT 1 FROM aggregated_signature WHERE signed_hash = $1);

-- name: InsertAggregatedSignature :execresult
INSERT INTO aggregated_signature (
    epoch_id, signed_hash, signers_bitfield, signature
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT DO NOTHING;

-- name: InsertEonPublicKey :exec
INSERT INTO eon_public_key (
    activation_block_number,
    eon_public_key
) VALUES (
    $1, $2
);

-- name: GetEonPublicKey :one
SELECT eon_public_key
FROM eon_public_key
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1;
