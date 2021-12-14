-- schema-version: 13 --
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
CREATE TABLE IF NOT EXISTS decryptor.aggregated_signature (
       epoch_id bytea,
       signed_hash bytea,
       signers_bitfield bytea PRIMARY KEY,
       signature bytea
);
CREATE TABLE IF NOT EXISTS decryptor.decryptor_identity (
       address text PRIMARY KEY,
       bls_public_key bytea,
       bls_signature bytea,
       signature_valid boolean NOT NULL
);
CREATE TABLE IF NOT EXISTS decryptor.decryptor_set_member (
       activation_block_number bigint NOT NULL,
       index int NOT NULL,
       address text NOT NULL,
       PRIMARY KEY (activation_block_number, index)
);
CREATE TABLE IF NOT EXISTS decryptor.keyper_set(
       activation_block_number bigint NOT NULL,
       keypers text[] NOT NULL,
       threshold integer NOT NULL
);
CREATE TABLE IF NOT EXISTS decryptor.eon_public_key (
       activation_block_number bigint PRIMARY KEY,
       eon_public_key bytea
);
CREATE TABLE IF NOT EXISTS decryptor.event_sync_progress (
       id bool UNIQUE NOT NULL DEFAULT true,
       next_block_number integer NOT NULL,
       next_log_index integer NOT NULL
);
INSERT INTO decryptor.event_sync_progress (next_block_number, next_log_index) VALUES (0,0);

CREATE TABLE decryptor.meta_inf(
       key text PRIMARY KEY,
       value text NOT NULL
);
CREATE TABLE decryptor.chain_keyper_set(
       n integer PRIMARY KEY,
       addresses text[] NOT NULL
);

CREATE TABLE decryptor.chain_collator(
       activation_block_number bigint PRIMARY KEY,
       collator text NOT NULL
);
