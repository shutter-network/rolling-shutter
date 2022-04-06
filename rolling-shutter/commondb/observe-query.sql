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
SELECT * FROM keyper_set
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1;

-- name: InsertChainCollator :exec
INSERT INTO chain_collator (activation_block_number, collator)
VALUES ($1, $2);

-- name: GetChainCollator :one
SELECT * FROM chain_collator
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1;
