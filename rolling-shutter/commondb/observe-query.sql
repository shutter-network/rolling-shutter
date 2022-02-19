-- name: UpdateEventSyncProgress :exec
INSERT INTO event_sync_progress (next_block_number, next_log_index)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
    SET next_block_number = $1,
        next_log_index = $2;

-- name: GetEventSyncProgress :one
SELECT next_block_number, next_log_index FROM event_sync_progress LIMIT 1;

-- name: GetNextBlockNumber :one
SELECT next_block_number from event_sync_progress LIMIT 1;

-- name: InsertKeyperSet :exec
INSERT INTO keyper_set (
    event_index,
    activation_block_number,
    keypers,
    threshold
) VALUES (
    $1, $2, $3, $4
);

-- name: GetKeyperSetByEventIndex :one
SELECT * FROM keyper_set WHERE event_index=$1;

-- name: GetKeyperSet :one
SELECT (
    activation_block_number,
    keypers,
    threshold
) FROM keyper_set
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1;

-- name: InsertDecryptorSetMember :exec
INSERT INTO decryptor_set_member (
    activation_block_number, index, address
) VALUES (
    $1, $2, $3
);

-- name: InsertDecryptorIdentity :exec
INSERT INTO decryptor_identity (
    address, bls_public_key, bls_signature, signature_valid
) VALUES (
    $1, $2, $3, $4
);

-- name: InsertChainCollator :exec
INSERT INTO chain_collator (activation_block_number, collator)
VALUES ($1, $2);

-- name: GetChainCollator :one
SELECT * FROM chain_collator
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1;

-- name: GetDecryptorIdentity :one
SELECT * FROM decryptor_identity
WHERE address = $1;
-- name: GetDecryptorSetMember :one
SELECT
    m1.activation_block_number,
    m1.index,
    m1.address,
    identity.bls_public_key,
    identity.bls_signature,
    coalesce(identity.signature_valid, false)
FROM (
    SELECT
        m2.activation_block_number,
        m2.index,
        m2.address
    FROM decryptor_set_member AS m2
    WHERE activation_block_number = (
        SELECT
            m3.activation_block_number
        FROM decryptor_set_member AS m3
        WHERE m3.activation_block_number <= $1
        ORDER BY m3.activation_block_number DESC
        LIMIT 1
    ) AND m2.index = $2
) AS m1
LEFT OUTER JOIN decryptor_identity AS identity
ON m1.address = identity.address
ORDER BY index;

-- name: GetDecryptorSet :many
SELECT
    member.activation_block_number,
    member.index,
    member.address,
    identity.bls_public_key,
    identity.bls_signature,
    coalesce(identity.signature_valid, false)
FROM (
    SELECT
        activation_block_number,
        index,
        address
    FROM decryptor_set_member
    WHERE activation_block_number = (
        SELECT
            m.activation_block_number
        FROM decryptor_set_member AS m
        WHERE m.activation_block_number <= $1
        ORDER BY m.activation_block_number DESC
        LIMIT 1
    )
) AS member
LEFT OUTER JOIN decryptor_identity AS identity
ON member.address = identity.address
ORDER BY index;
