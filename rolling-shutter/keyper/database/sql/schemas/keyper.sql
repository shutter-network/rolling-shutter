-- schema-version: keyper-16 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE TABLE decryption_trigger (
       epoch_id bytea PRIMARY KEY
);
CREATE TABLE decryption_key_share (
       eon bigint,
       epoch_id bytea,
       keyper_index bigint,
       decryption_key_share bytea,
       PRIMARY KEY (eon, epoch_id, keyper_index)
);
CREATE TABLE decryption_key (
       eon bigint,
       epoch_id bytea,
       decryption_key bytea,
       PRIMARY KEY (eon, epoch_id)
);

----- legacy tendermint tables. Dropped in migration V5.

CREATE TABLE last_batch_config_sent(
       enforce_one_row BOOL PRIMARY KEY DEFAULT TRUE,
       keyper_config_index bigint NOT NULL
);
INSERT INTO last_batch_config_sent (keyper_config_index) VALUES (0);

CREATE TABLE last_block_seen(
       enforce_one_row BOOL PRIMARY KEY DEFAULT TRUE,
       block_number bigint NOT NULL
);
INSERT INTO last_block_seen (block_number) VALUES (-1);

CREATE TABLE tendermint_sync_meta (
       current_block bigint NOT NULL,
       last_committed_height bigint NOT NULL,
       sync_timestamp timestamp NOT NULL,
       PRIMARY KEY (current_block, last_committed_height)
);

CREATE TABLE puredkg (
       eon bigint PRIMARY KEY,
       puredkg BYTEA NOT NULL
);

CREATE TABLE tendermint_batch_config(
       keyper_config_index integer PRIMARY KEY,
       height bigint NOT NULL,
       keypers text[] NOT NULL,
       threshold integer NOT NULL,
       started boolean NOT NULL,
       activation_block_number bigint NOT NULL
);

CREATE TABLE tendermint_encryption_key(
       address TEXT PRIMARY KEY,
       encryption_public_key BYTEA NOT NULL
);

CREATE TABLE tendermint_outgoing_messages(
       id SERIAL PRIMARY KEY,
       description TEXT NOT NULL,
       msg BYTEA NOT NULL
);

CREATE TABLE eons(
       eon bigint PRIMARY KEY,
       height bigint NOT NULL,
       activation_block_number bigint NOT NULL,
       keyper_config_index bigint NOT NULL
);

CREATE TABLE poly_evals(
       eon bigint NOT NULL,
       receiver_address TEXT NOT NULL,
       eval BYTEA NOT NULL,
       PRIMARY KEY (eon, receiver_address)
);

-- dkg_result contains the result of running the DKG process or an error message, if the DKG
-- process failed.
CREATE TABLE dkg_result(
       eon bigint PRIMARY KEY,
       success BOOLEAN NOT NULL,
       error TEXT,
       pure_result BYTEA  -- shdb.EncodePureDKGResult/shdb.DecodePureDKGResult
);

-- outgoing_eon_keys contains the eon public key(s) that should be broadcast as a result of a successful DKG
CREATE TABLE outgoing_eon_keys(
       eon_public_key bytea,
       eon bigint NOT NULL PRIMARY KEY
);
