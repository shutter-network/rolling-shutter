-- schema-version: 2 --
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
