-- schema-version: 7 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.


CREATE SCHEMA IF NOT EXISTS decryptor;

CREATE TABLE IF NOT EXISTS decryptor.cipher_batch (
       epoch_id bytea PRIMARY KEY,
       transactions bytea[]
);
CREATE TABLE IF NOT EXISTS decryptor.decryption_key (
       epoch_id bytea PRIMARY KEY,
       key bytea
);
CREATE TABLE IF NOT EXISTS decryptor.decryption_signature (
       epoch_id bytea,
       signed_hash bytea,
       signers_bitfield bytea,
       signature bytea,
       PRIMARY KEY (epoch_id, signers_bitfield)
);
CREATE TABLE IF NOT EXISTS decryptor.decryptor_identity (
       address text PRIMARY KEY,
       bls_public_key bytea
);
CREATE TABLE IF NOT EXISTS decryptor.decryptor_set_member (
       start_epoch_id bytea,
       index int,
       address text NOT NULL,
       PRIMARY KEY (start_epoch_id, index)
);
CREATE TABLE IF NOT EXISTS decryptor.keyper_set(
       start_epoch_id bytea,
       keypers text[] NOT NULL,
       threshold integer NOT NULL
);
CREATE TABLE IF NOT EXISTS decryptor.eon_public_key (
       start_epoch_id bytea PRIMARY KEY,
       eon_public_key bytea
);
CREATE TABLE decryptor.meta_inf(
       key text PRIMARY KEY,
       value text NOT NULL
);
