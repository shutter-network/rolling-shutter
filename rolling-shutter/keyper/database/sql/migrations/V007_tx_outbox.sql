-- schema-version: keyper-21 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

CREATE TABLE tx_outbox (
    id           BIGSERIAL PRIMARY KEY,
    to_address   TEXT        NOT NULL,
    data         BYTEA       NOT NULL,
    value        NUMERIC     NOT NULL DEFAULT 0,
    status       TEXT        NOT NULL DEFAULT 'pending',
    tx_hash      TEXT,
    nonce        BIGINT,
    error        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ
);

CREATE INDEX tx_outbox_status_idx ON tx_outbox (status, id);
