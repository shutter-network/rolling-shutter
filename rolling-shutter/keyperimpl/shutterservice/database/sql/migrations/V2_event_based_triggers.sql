CREATE TABLE event_trigger_registered_event (
    block_number bigint NOT NULL CHECK (block_number >= 0),
    block_hash bytea NOT NULL,
    tx_index bigint NOT NULL CHECK (tx_index >= 0),
    log_index bigint NOT NULL CHECK (log_index >= 0),
    eon bigint NOT NULL CHECK (eon >= 0),
    identity_prefix bytea NOT NULL,
    sender text NOT NULL,
    definition bytea NOT NULL,
    expiration_block_number bigint NOT NULL CHECK (expiration_block_number >= 0),
    decrypted boolean NOT NULL DEFAULT false,
    identity bytea NOT NULL,
    PRIMARY KEY (eon, identity_prefix, sender)
);

CREATE TABLE multi_event_sync_status (
    enforce_one_row bool PRIMARY KEY DEFAULT true,
    block_number bigint NOT NULL CHECK (block_number >= 0),
    block_hash bytea NOT NULL
);

CREATE TABLE fired_triggers (
    eon bigint NOT NULL,
    identity_prefix bytea NOT NULL,
    sender text NOT NULL,
    block_number bigint NOT NULL CHECK (block_number >= 0),
    block_hash bytea NOT NULL,
    tx_index bigint NOT NULL CHECK (tx_index >= 0),
    log_index bigint NOT NULL CHECK (log_index >= 0),
    PRIMARY KEY (eon, identity_prefix, sender),
    FOREIGN KEY (eon, identity_prefix, sender) REFERENCES event_trigger_registered_event (eon, identity_prefix, sender) ON DELETE CASCADE
);