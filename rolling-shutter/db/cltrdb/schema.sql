-- schema-version: collator-12 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE TABLE decryption_trigger(
    epoch_id bytea PRIMARY KEY,
    -- id persists the input ordering of trigger
    -- this is useful for implementing a message send queue
    -- since the epoch_id does not have to be incremental
    id INTEGER GENERATED ALWAYS AS IDENTITY NOT NULL,
    batch_hash bytea,
    l1_block_number bigint NOT NULL,
    sent timestamp
);

CREATE INDEX unsent_decryption_trigger_idx
ON decryption_trigger((sent IS NULL)) WHERE (sent IS NULL);

CREATE OR REPLACE FUNCTION notify_new_decryption_trigger()
  RETURNS TRIGGER AS $$
DECLARE
BEGIN
  PERFORM pg_notify('new_decryption_trigger', 'payload');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_decryption_trigger
         AFTER INSERT ON decryption_trigger
    FOR EACH STATEMENT EXECUTE PROCEDURE notify_new_decryption_trigger();

CREATE TABLE decryption_key (
       epoch_id bytea PRIMARY KEY,
       decryption_key bytea
);

CREATE OR REPLACE FUNCTION notify_new_decryption_key()
  RETURNS TRIGGER AS $$
DECLARE
BEGIN
  PERFORM pg_notify('new_decryption_key', 'payload');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_decryption_key
        AFTER INSERT ON decryption_key
    FOR EACH STATEMENT EXECUTE PROCEDURE notify_new_decryption_key();



CREATE TYPE txstatus AS ENUM ('new', 'rejected', 'committed');

CREATE TABLE transaction(
       tx_hash bytea PRIMARY KEY,
       -- id persists the input ordering of txs
       id INTEGER GENERATED ALWAYS AS IDENTITY,
       epoch_id bytea,
       tx_bytes bytea,
       status txstatus NOT NULL
       );

-- next_batch contains data to be used in the next batch to be submitted. It will be populated
-- as soon as the previous batch has been finalized.
CREATE TABLE next_batch(
    enforce_one_row BOOL PRIMARY KEY DEFAULT TRUE,
    epoch_id bytea NOT NULL,
    l1_block_number bigint NOT NULL
);

CREATE TABLE batchtx(
       epoch_id bytea PRIMARY KEY,
       marshaled bytea NOT NULL,
       submitted BOOL DEFAULT FALSE NOT NULL
);

CREATE OR REPLACE FUNCTION notify_new_batchtx()
  RETURNS TRIGGER AS $$
DECLARE
BEGIN
  PERFORM pg_notify('new_batchtx', 'payload');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_batchtx
    AFTER INSERT ON batchtx
    FOR EACH STATEMENT EXECUTE PROCEDURE notify_new_batchtx();

-- ensure we only have at most one tx not submitted yet
CREATE UNIQUE INDEX batchtx_at_most_one_not_yet_submitted ON batchtx (submitted) WHERE submitted = false;

-- CREATE TABLE eon(
--      activation_block_number bigint NOT NULL,
--      eon_public_key bytea,
--      threshold bigint NOT NULL,
--      PRIMARY KEY (eon_public_key, activation_block_number)
-- );

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
