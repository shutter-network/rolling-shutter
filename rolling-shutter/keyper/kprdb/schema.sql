-- schema-version: 7 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE SCHEMA keyper;
CREATE TABLE keyper.decryption_trigger (
       epoch_id bytea PRIMARY KEY
);
CREATE TABLE keyper.decryption_key_share (
       epoch_id bytea,
       keyper_index bigint,
       decryption_key_share bytea,
       PRIMARY KEY (epoch_id, keyper_index)
);
CREATE TABLE keyper.decryption_key (
       epoch_id bytea PRIMARY KEY,
       keyper_index bigint,
       decryption_key bytea
);
CREATE TABLE keyper.meta_inf(
       key text PRIMARY KEY,
       value text NOT NULL
);

----- tendermint events

-- tendermint_sync_meta contains meta information about the synchronization process with the
-- tendermint app. At the moment we just insert new entries into the table and sort by
-- current_block to get the latest entry. When handling new events from shuttermint, we do that in
-- batches inside a PostgreSQL transaction. last_committed_height is the last block that we know is
-- available, current_block is the last block in the batch we're currently handling.
CREATE TABLE keyper.tendermint_sync_meta (
       current_block bigint NOT NULL,
       last_committed_height bigint NOT NULL,
       sync_timestamp timestamp NOT NULL,
       PRIMARY KEY (current_block, last_committed_height)
);

-- keyper.puredkg contains a gob serialized puredkg instance.  We already have the DKG process
-- implemented in go, without any database access.  When new events come in, we feed those to the
-- go object and store it afterwards in the puredkg table.
CREATE TABLE keyper.puredkg (
       eon bigint PRIMARY KEY,
       puredkg BYTEA NOT NULL
);

CREATE TABLE keyper.tendermint_batch_config(
       config_index integer PRIMARY KEY,
       height bigint NOT NULL,
       keypers text[] NOT NULL,
       threshold integer NOT NULL
);

CREATE TABLE keyper.tendermint_encryption_key(
       address TEXT PRIMARY KEY,
       encryption_public_key BYTEA NOT NULL
);

CREATE TABLE keyper.tendermint_outgoing_messages(
       id SERIAL PRIMARY KEY,
       description TEXT NOT NULL,
       msg BYTEA NOT NULL
);

CREATE TABLE keyper.eons(
       eon bigint PRIMARY KEY,
       height bigint NOT NULL,
       batch_index BYTEA NOT NULL,
       config_index bigint NOT NULL
);

CREATE TABLE keyper.poly_evals(
       eon bigint NOT NULL,
       receiver_address TEXT NOT NULL,
       eval BYTEA NOT NULL,
       PRIMARY KEY (eon, receiver_address)
);

-- dkg_result contains the result of running the DKG process or an error message, if the DKG
-- process failed.
CREATE TABLE keyper.dkg_result(
       eon bigint PRIMARY KEY,
       success BOOLEAN NOT NULL,
       error TEXT,
       pure_result BYTEA  -- shdb.EncodePureDKGResult/shdb.DecodePureDKGResult
);
