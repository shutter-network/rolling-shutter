-- schema-version: collator-8 --
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
