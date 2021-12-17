-- add a decryptor to the decryptor set starting at epoch 0

INSERT INTO decryptor_identity (
    address,
    bls_public_key
) VALUES (
    :address,
    :key
) ON CONFLICT DO NOTHING;

INSERT INTO decryptor_set_member (
    activation_block_number,
    index,
    address
) VALUES (
    0,
    (
        SELECT
            count(1)
        FROM decryptor_set_member
        WHERE activation_block_number = 0
    ),
    :address
);
