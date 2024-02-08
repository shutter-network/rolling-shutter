-- schema-version: gnosiskeyper-1 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE TABLE transaction_submitted_event (
    index bigint PRIMARY KEY CHECK (index >= 0),
    block_number bigint NOT NULL CHECK (block_number >= 0),
    block_hash bytea NOT NULL,
    tx_index bigint NOT NULL CHECK (tx_index >= 0),
    log_index bigint NOT NULL CHECK (log_index >= 0),
    eon bigint NOT NULL CHECK (eon >= 0),
    identity_prefix bytea NOT NULL,
    sender text NOT NULL,
    gas_limit bigint NOT NULL CHECK (gas_limit >= 0)
);

CREATE TABLE transaction_submitted_events_synced_until(
    enforce_one_row bool PRIMARY KEY DEFAULT true,
    block_number bigint NOT NULL CHECK (block_number >= 0)
);

CREATE TABLE transaction_submitted_event_count(
    eon bigint PRIMARY KEY,
    event_count bigint NOT NULL DEFAULT 0 CHECK (event_count >= 0)
);

-- tx_pointer stores what we know about the current value of the tx pointer. There are two
-- sources that are stored independently: The keyper itself (local) and the other keypers
-- (consensus). The local value is updated whenever the keyper sends decryption key shares for some
-- transactions. The consensus value is updated when decryption keys are received from other
-- keypers. The values can be NULL if the keyper has lost track (e.g. after a restart or if no
-- messages have been received recently). Each value is annotated with the number of the block at
-- whose end the tx_pointer is valid.
CREATE TABLE tx_pointer(
    enforce_one_row bool PRIMARY KEY DEFAULT true,
    local bigint DEFAULT NULL,
    local_block bigint NOT NULL DEFAULT 0,
    consensus bigint DEFAULT NULL,
    consensus_block bigint NOT NULL DEFAULT 0
);