-- name: GetCommitmentByTxHash :many
SELECT
    c.tx_hashes,
    c.provider_address,
    c.commitment_signature,
    c.commitment_digest,
    c.block_number,
    c.received_bid_digest,
    c.received_bid_signature,
    c.bidder_node_address
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
    ON CONFLICT (eon, identity_preimage, tx_hash)
    DO UPDATE SET
        block_number = EXCLUDED.block_number
    RETURNING tx_hash as hashes
),
upserted_commitment AS (
    INSERT INTO commitment (tx_hashes, provider_address, commitment_signature, commitment_digest, block_number, received_bid_digest, received_bid_signature, bidder_node_address)
    SELECT
        ARRAY_AGG(hashes),
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11
    FROM inserted_transactions
    ON CONFLICT (provider_address, commitment_digest, block_number)
    DO UPDATE SET
        tx_hashes = commitment.tx_hashes || EXCLUDED.tx_hashes,
        received_bid_digest = EXCLUDED.received_bid_digest,
        received_bid_signature = EXCLUDED.received_bid_signature,
        bidder_node_address = EXCLUDED.bidder_node_address
    RETURNING tx_hashes, provider_address
)
SELECT tx_hashes, provider_address FROM upserted_commitment;

-- name: GetProviderRegistryEventsSyncedUntil :one
SELECT * FROM provider_registry_events_synced_until LIMIT 1;

-- name: SetProviderRegistryEventsSyncedUntil :exec
INSERT INTO provider_registry_events_synced_until (block_hash, block_number) VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_hash = $1, block_number = $2;

-- name: InsertProviderRegistryEvent :execresult
INSERT INTO provider_registry_events (block_number, block_hash, tx_index, log_index, provider_address, bls_keys)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (provider_address) DO UPDATE SET
block_number = $1,
block_hash = $2,
tx_index = $3,
log_index = $4,
bls_keys = $6;

-- name: DeleteProviderRegistryEventsFromBlockNumber :exec
DELETE FROM provider_registry_events WHERE block_number >= $1;