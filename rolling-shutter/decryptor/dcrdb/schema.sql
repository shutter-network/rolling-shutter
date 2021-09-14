-- schema-version: 2 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.


CREATE SCHEMA IF NOT EXISTS decryptor;

CREATE TABLE IF NOT EXISTS decryptor.cipher_batch (
       epoch_id bytea PRIMARY KEY,
       data bytea
);
CREATE TABLE IF NOT EXISTS decryptor.decryption_key (
       epoch_id bytea PRIMARY KEY,
       key bytea
);
CREATE TABLE IF NOT EXISTS decryptor.decryption_signature (
       epoch_id bytea,
       signed_hash bytea,
       signer_index bigint,
       signature bytea,
       PRIMARY KEY (epoch_id, signer_index)
);
CREATE TABLE decryptor.meta_inf(
       key text PRIMARY KEY,
       value text NOT NULL
);
