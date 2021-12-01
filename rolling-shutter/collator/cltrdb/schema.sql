-- schema-version: 2 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.


CREATE SCHEMA IF NOT EXISTS collator;

CREATE TABLE collator.meta_inf(
       key text PRIMARY KEY,
       value text NOT NULL
);

CREATE TABLE collator.decryption_trigger(
    epoch_id bytea PRIMARY KEY,
    batch_hash bytea
);

CREATE TABLE collator.cipher_batch(
    epoch_id bytea PRIMARY KEY,
    transactions bytea[]
);

CREATE TABLE collator.transaction(
       tx_id bytea PRIMARY KEY,
       epoch_id bytea,
       encrypted_tx bytea
);
