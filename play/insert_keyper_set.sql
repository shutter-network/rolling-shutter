-- name: InsertKeyperSet :exec
INSERT INTO keyper_set (
    activation_block_number,
    keypers,
    threshold
) VALUES (
    0,
    :keypers,
    1
);
