-- name: InsertKeyperSet :exec
INSERT INTO keyper_set (
    keyper_config_index,
    activation_block_number,
    keypers,
    threshold
) VALUES (
    $1, $2, $3, $4
);

-- name: GetKeyperSetByKeyperConfigIndex :one
SELECT * FROM keyper_set WHERE keyper_config_index=$1;

-- name: GetKeyperSet :one
SELECT * FROM keyper_set
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1;

-- name: GetKeyperSets :many
SELECT * FROM keyper_set
ORDER BY activation_block_number ASC;