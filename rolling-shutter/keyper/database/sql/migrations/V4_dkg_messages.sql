-- schema-version: keyper-18 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

CREATE TABLE dkg_poly_commitment (
    keyper_config_index  bigint NOT NULL,
    retry_counter        bigint NOT NULL,
    keyper_index         bigint NOT NULL,
    commitment           bytea  NOT NULL,
    PRIMARY KEY (keyper_config_index, retry_counter, keyper_index)
);

CREATE TABLE dkg_poly_eval (
    keyper_config_index  bigint NOT NULL,
    retry_counter        bigint NOT NULL,
    sender_index         bigint NOT NULL,
    receiver_index       bigint NOT NULL,
    encrypted_eval       bytea  NOT NULL,
    PRIMARY KEY (keyper_config_index, retry_counter, sender_index, receiver_index)
);

CREATE TABLE dkg_accusation (
    keyper_config_index  bigint NOT NULL,
    retry_counter        bigint NOT NULL,
    accuser_index        bigint NOT NULL,
    accused_index        bigint NOT NULL,
    PRIMARY KEY (keyper_config_index, retry_counter, accuser_index, accused_index)
);

CREATE TABLE dkg_apology (
    keyper_config_index  bigint NOT NULL,
    retry_counter        bigint NOT NULL,
    apologizer_index     bigint NOT NULL,
    accuser_index        bigint NOT NULL,
    poly_eval            bytea  NOT NULL,
    PRIMARY KEY (keyper_config_index, retry_counter, apologizer_index, accuser_index)
);

ALTER TABLE eons DROP COLUMN height;
