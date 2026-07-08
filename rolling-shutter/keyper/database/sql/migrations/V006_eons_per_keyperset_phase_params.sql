-- schema-version: keyper-20 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

ALTER TABLE eons
    ADD COLUMN dkg_contract TEXT,
    ADD COLUMN phase_length BIGINT,
    ADD COLUMN lead_length  BIGINT;
