-- name: InsertKeyperSet :exec
INSERT INTO decryptor.keyper_set (
    activation_block_number,
    keypers,
    threshold
) VALUES (
    0,
    :keypers,
    1
);
