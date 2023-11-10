CREATE TABLE IF NOT EXISTS decryption_key (
        epoch_id bytea PRIMARY KEY,
        key bytea
);
CREATE TABLE IF NOT EXISTS eon_public_key (
        eon_id bigint PRIMARY KEY,
        eon_public_key bytea
);
