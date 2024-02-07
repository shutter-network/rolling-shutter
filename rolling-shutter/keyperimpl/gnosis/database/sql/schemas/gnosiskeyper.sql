-- schema-version: gnosiskeyper-1 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE TABLE transaction_submitted_event (
    block_number bigint CHECK (block_number >= 0),
    block_hash bytea,
    tx_index bigint CHECK (tx_index >= 0),
    log_index bigint CHECK (log_index >= 0),
    eon bigint NOT NULL CHECK (eon >= 0),
    identity_prefix bytea NOT NULL,
    sender text NOT NULL,
    gas_limit bigint NOT NULL CHECK (gas_limit >= 0),
    PRIMARY KEY (block_number, block_hash, tx_index, log_index)
);

create table transaction_submitted_events_synced_until(
       enforce_one_row bool primary key default true,
       block_number bigint not null CHECK (block_number > 0)
);