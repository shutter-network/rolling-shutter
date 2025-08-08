-- schema-version: gnosiskeyper-2 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

ALTER TABLE validator_registrations
 DROP CONSTRAINT validator_registrations_pkey,
 ADD PRIMARY KEY (block_number, tx_index, log_index, validator_index);