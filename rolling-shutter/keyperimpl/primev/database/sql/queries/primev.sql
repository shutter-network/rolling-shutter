-- name: GetCommitmentByTxHash :many
SELECT
    c.tx_hashes,
    c.provider_address,
    c.commitment_signature,
    c.commitment_digest,
    c.block_number
FROM commitment c
WHERE c.tx_hashes = $1;

-- name: InsertMultipleTransactionsAndUpsertCommitment :exec
WITH inserted_transactions AS (
    INSERT INTO committed_transactions (eon, identity_preimage, block_number, tx_hash)
    SELECT
        unnest($1::bigint[]) as eon,
        unnest($2::text[]) as identity_preimage,
        unnest($3::bigint[]) as block_number,
        unnest($4::text[]) as tx_hash
    ON CONFLICT (eon, identity_preimage, tx_hash) DO NOTHING
    RETURNING tx_hash
),
upserted_commitment AS (
    INSERT INTO commitment (tx_hashes, provider_address, commitment_signature, commitment_digest, block_number)
    SELECT
        ARRAY_AGG(tx_hash),
        $5,
        $6,
        $7,
        $8
    FROM inserted_transactions
    ON CONFLICT (provider_address, commitment_digest, block_number)
    DO UPDATE SET
        tx_hashes = commitment.tx_hashes || EXCLUDED.tx_hashes
    RETURNING tx_hashes, provider_address
)
SELECT tx_hashes, provider_address FROM upserted_commitment;

