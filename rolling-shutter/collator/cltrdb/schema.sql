-- schema-version: collator-8 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE TABLE decryption_trigger(
    epoch_id bytea PRIMARY KEY,
    batch_hash bytea
);

CREATE TABLE decryption_key (
       epoch_id bytea PRIMARY KEY,
       decryption_key bytea
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

CREATE TABLE eon_public_key_candidate(
    hash bytea PRIMARY KEY,
    eon_public_key bytea NOT NULL,
    activation_block_number bigint NOT NULL,
    keyper_config_index bigint NOT NULL,
    eon bigint NOT NULL,
    confirmed BOOL NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX eon_public_key_index ON eon_public_key_candidate(eon_public_key, activation_block_number, keyper_config_index, eon);

-- eon_public_key_vote stores the votes. This maps a sender address to a hash. The eon and
-- keyper_config_index fields are only here to create unique indexes on them, since postgresql does
-- not allow us to create indexes on views. They will match the values referenced in the
-- eon_public_key_candidate table.
CREATE TABLE eon_public_key_vote(
    hash bytea REFERENCES eon_public_key_candidate(hash),
    sender text NOT NULL,
    signature bytea NOT NULL,
    eon bigint NOT NULL,
    keyper_config_index bigint NOT NULL,
    PRIMARY KEY(sender, eon)
);

-- allow each sender to vote for at most one keyper_config_index.
CREATE UNIQUE INDEX eon_public_key_votes_unique_per_keyper_config_index ON eon_public_key_vote(sender, keyper_config_index);
