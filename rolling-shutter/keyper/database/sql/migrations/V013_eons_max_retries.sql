-- schema-version: keyper-27 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

ALTER TABLE eons
    ADD COLUMN max_retries BIGINT NOT NULL;
