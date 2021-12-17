-- schema-version: decryptor-14 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.


CREATE TABLE IF NOT EXISTS cipher_batch (
       epoch_id bytea PRIMARY KEY,
       transactions bytea[]
);
CREATE TABLE IF NOT EXISTS decryption_key (
       epoch_id bytea PRIMARY KEY,
       key bytea
);
CREATE TABLE IF NOT EXISTS decryption_signature (
       epoch_id bytea,
       signed_hash bytea,
       signers_bitfield bytea,
       signature bytea,
       PRIMARY KEY (epoch_id, signers_bitfield)
);
CREATE TABLE IF NOT EXISTS aggregated_signature (
       epoch_id bytea,
       signed_hash bytea,
       signers_bitfield bytea PRIMARY KEY,
       signature bytea
);
CREATE TABLE IF NOT EXISTS eon_public_key (
       activation_block_number bigint PRIMARY KEY,
       eon_public_key bytea
);
