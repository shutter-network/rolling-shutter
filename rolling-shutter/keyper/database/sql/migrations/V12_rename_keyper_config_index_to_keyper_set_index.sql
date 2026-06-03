-- schema-version: keyper-26 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

ALTER TABLE dkg_poly_commitments RENAME COLUMN keyper_config_index TO keyper_set_index;
ALTER TABLE dkg_poly_evals       RENAME COLUMN keyper_config_index TO keyper_set_index;
ALTER TABLE dkg_accusations      RENAME COLUMN keyper_config_index TO keyper_set_index;
ALTER TABLE dkg_apologies        RENAME COLUMN keyper_config_index TO keyper_set_index;
ALTER TABLE dkg_initial_states   RENAME COLUMN keyper_config_index TO keyper_set_index;
ALTER TABLE dkg_sent_actions     RENAME COLUMN keyper_config_index TO keyper_set_index;
