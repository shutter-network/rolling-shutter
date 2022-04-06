-- schema-version: collator-7 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE TABLE decryption_trigger(
    epoch_id bytea PRIMARY KEY,
    batch_hash bytea
);

CREATE TABLE transaction(
       tx_id bytea PRIMARY KEY,
       epoch_id bytea,
       encrypted_tx bytea
);

CREATE TABLE next_epoch(
    enforce_one_row BOOL PRIMARY KEY DEFAULT TRUE,
    epoch_id bytea NOT NULL
);

CREATE TABLE eon(
     activation_block_number bigint NOT NULL,
     eon_public_key bytea,
     threshold bigint NOT NULL,
     PRIMARY KEY (eon_public_key, activation_block_number)
);

-- those will be only be inserted when they are valid (signature etc)
CREATE TABLE eon_public_key_message(
    eon_public_key bytea,
    activation_block_number bigint NOT NULL,
    keyper_index bigint NOT NULL,
    msg_bytes bytea,
    PRIMARY KEY (eon_public_key, activation_block_number, keyper_index)
);
