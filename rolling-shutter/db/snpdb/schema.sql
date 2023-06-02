-- schema-version: snapshot-1 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.


CREATE TABLE IF NOT EXISTS decryption_key (
        epoch_id bytea PRIMARY KEY,
        key bytea
);
CREATE TABLE IF NOT EXISTS eon_public_key (
        eon_id bigint PRIMARY KEY,
        eon_public_key bytea
);
