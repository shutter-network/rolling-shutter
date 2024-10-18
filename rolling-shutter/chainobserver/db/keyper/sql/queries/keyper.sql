-- name: InsertKeyperSet :exec
INSERT INTO keyper_set (
    keyper_config_index,
    activation_block_number,
    keypers,
    threshold
) VALUES (
    $1, $2, $3, $4
) ON CONFLICT DO NOTHING;

-- name: GetKeyperSetByKeyperConfigIndex :one
SELECT * FROM keyper_set WHERE keyper_config_index=$1;

-- name: GetKeyperSet :one
SELECT * FROM keyper_set
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1;

-- name: GetKeyperSets :many
SELECT * FROM keyper_set
ORDER BY activation_block_number ASC;

-- name: InsertSyncedBlock :exec
INSERT INTO recent_block (
       block_hash,
       block_number,
       parent_hash,
       timestamp,
       header
) VALUES (
    $1, $2, $3, $4, $5
) ON CONFLICT DO UPDATE SET
       block_hash = $1,
       block_number =$2 ,
       parent_hash =$3,
       timestamp =$4 ,
       header =$5
;

-- name: GetSyncedBlockByHash :one
SELECT * FROM recent_block
WHERE block_hash = $1;

-- name: DeleteSyncedBlockByHash :exec
DELETE FROM recent_block
WHERE block_hash = $1;

-- name: GetSyncedBlocks :many
SELECT * FROM recent_block
ORDER BY block_number DESC;

-- name: GetLatestSyncedBlocks :many
SELECT * FROM recent_block
ORDER BY block_number DESC
LIMIT $1;

-- name: EvictSyncedBlocksBefore :exec
DELETE FROM recent_block
WHERE block_number < $1;
