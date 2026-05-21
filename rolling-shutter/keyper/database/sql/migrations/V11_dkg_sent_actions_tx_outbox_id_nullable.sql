-- schema-version: keyper-25 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

ALTER TABLE dkg_sent_actions RENAME COLUMN outbox_id TO tx_outbox_id;
ALTER TABLE dkg_sent_actions ALTER COLUMN tx_outbox_id DROP NOT NULL;
