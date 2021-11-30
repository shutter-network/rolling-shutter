-- insert an eon public key with start epoch 0

INSERT INTO decryptor.eon_public_key (
    activation_block_number,
    eon_public_key
) VALUES (
    0,
    :eon_public_key
);
