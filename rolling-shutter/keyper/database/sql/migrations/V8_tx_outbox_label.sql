-- schema-version: keyper-22 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

ALTER TABLE tx_outbox
    ADD COLUMN label TEXT NOT NULL DEFAULT '';
