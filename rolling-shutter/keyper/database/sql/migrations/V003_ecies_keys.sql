-- schema-version: keyper-17 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

CREATE TABLE ecies_keys (
    keyper_address    text  NOT NULL PRIMARY KEY,
    ecies_public_key  bytea NOT NULL
);
