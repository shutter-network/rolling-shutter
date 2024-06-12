-- schema-version: gnosiskeyper-1 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE TABLE transaction_submitted_event (
    index bigint CHECK (index >= 0),
    block_number bigint NOT NULL CHECK (block_number >= 0),
    block_hash bytea NOT NULL,
    tx_index bigint NOT NULL CHECK (tx_index >= 0),
    log_index bigint NOT NULL CHECK (log_index >= 0),
    eon bigint NOT NULL CHECK (eon >= 0),
    identity_prefix bytea NOT NULL,
    sender text NOT NULL,
    gas_limit bigint NOT NULL CHECK (gas_limit >= 0),
    PRIMARY KEY (index, eon)
);

CREATE TABLE transaction_submitted_events_synced_until(
    enforce_one_row bool PRIMARY KEY DEFAULT true,
    block_hash bytea NOT NULL,
    block_number bigint NOT NULL CHECK (block_number >= 0),
    slot bigint NOT NULL CHECK (slot >= 0)
);

CREATE TABLE transaction_submitted_event_count(
    eon bigint PRIMARY KEY,
    event_count bigint NOT NULL DEFAULT 0 CHECK (event_count >= 0)
);

CREATE TABLE tx_pointer(
    eon bigint PRIMARY KEY,
    slot bigint NOT NULL DEFAULT 0,
    value bigint NOT NULL DEFAULT 0
);

CREATE TABLE current_decryption_trigger(
    eon bigint PRIMARY KEY CHECK (eon >= 0),
    slot bigint NOT NULL CHECK (slot >= 0),
    tx_pointer bigint NOT NULL CHECK (tx_pointer >= 0),
    identities_hash bytea NOT NULL
);

CREATE TABLE slot_decryption_signatures(
    eon bigint NOT NULL CHECK (eon >= 0),
    slot bigint NOT NULL CHECK (slot >= 0),
    keyper_index bigint NOT NULL,
    tx_pointer bigint NOT NULL CHECK (tx_pointer >= 0),
    identities_hash bytea NOT NULL,
    signature bytea NOT NULL,
    PRIMARY KEY (eon, slot, keyper_index)
);

CREATE TABLE validator_registrations(
    block_number bigint NOT NULL CHECK (block_number >= 0),
    block_hash bytea NOT NULL,
    tx_index bigint NOT NULL CHECK (tx_index >= 0),
    log_index bigint NOT NULL CHECK (log_index >= 0),
    validator_index bigint NOT NULL CHECK (validator_index >= 0),
    nonce bigint NOT NULL CHECK (nonce >= 0),
    is_registration bool NOT NULL,
    PRIMARY KEY (block_number, tx_index, log_index)
);
CREATE INDEX idx_validator_index ON validator_registrations (validator_index);

CREATE TABLE validator_registrations_synced_until(
    enforce_one_row bool PRIMARY KEY DEFAULT true,
    block_hash bytea NOT NULL,
    block_number bigint NOT NULL CHECK (block_number >= 0)
);