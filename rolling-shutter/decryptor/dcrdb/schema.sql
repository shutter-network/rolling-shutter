CREATE SCHEMA IF NOT EXISTS decryptor;

CREATE TABLE IF NOT EXISTS decryptor.cipher_batch (
       epoch_id bigint PRIMARY KEY,
       data bytea
);
CREATE TABLE IF NOT EXISTS decryptor.decryption_key (
       epoch_id bigint PRIMARY KEY,
       key bytea
);
CREATE TABLE IF NOT EXISTS decryptor.decryption_signature (
       epoch_id bigint,
       signed_hash bytea,
       signer_index bigint,
       signature bytea,
       PRIMARY KEY (epoch_id, signer_index)
);
