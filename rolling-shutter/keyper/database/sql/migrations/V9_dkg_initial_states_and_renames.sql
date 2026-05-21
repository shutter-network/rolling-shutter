-- schema-version: keyper-23 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

ALTER TABLE dkg_poly_commitment RENAME TO dkg_poly_commitments;
ALTER TABLE dkg_poly_eval       RENAME TO dkg_poly_evals;
ALTER TABLE dkg_accusation      RENAME TO dkg_accusations;
ALTER TABLE dkg_apology         RENAME TO dkg_apologies;

CREATE TABLE dkg_initial_states (
    keyper_config_index bigint NOT NULL,
    retry_counter       bigint NOT NULL,
    puredkg_bytes       bytea  NOT NULL,
    PRIMARY KEY (keyper_config_index, retry_counter)
);
