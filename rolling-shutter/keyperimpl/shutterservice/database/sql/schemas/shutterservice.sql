-- schema-version: shutterservice-1 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE TABLE identity_registered_event (
    index bigint CHECK (index >= 0),
    block_number bigint NOT NULL CHECK (block_number >= 0),
    block_hash bytea NOT NULL,
    tx_index bigint NOT NULL CHECK (tx_index >= 0),
    log_index bigint NOT NULL CHECK (log_index >= 0),
    eon bigint NOT NULL CHECK (eon >= 0),
    identity_prefix bytea NOT NULL,
    sender text NOT NULL,
    timestamp bigint NOT NULL,
    decrypted boolean NOT NULL,
    PRIMARY KEY (index, eon)
);

CREATE TABLE identity_registered_events_synced_until(
    enforce_one_row bool PRIMARY KEY DEFAULT true,
    block_hash bytea NOT NULL,
    block_number bigint NOT NULL CHECK (block_number >= 0)
);

CREATE TABLE current_decryption_trigger(
    eon bigint PRIMARY KEY CHECK (eon >= 0),
    last_block_number bigint NOT NULL CHECK (last_block_number >= 0),
    identities_hash bytea NOT NULL
);
