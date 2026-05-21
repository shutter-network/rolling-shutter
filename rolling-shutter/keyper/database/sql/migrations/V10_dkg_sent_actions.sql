-- schema-version: keyper-24 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

CREATE TABLE dkg_sent_actions (
    keyper_config_index bigint NOT NULL,
    retry_counter       bigint NOT NULL,
    action              text   NOT NULL,
    outbox_id           bigint NOT NULL REFERENCES tx_outbox(id),
    PRIMARY KEY (keyper_config_index, retry_counter, action)
);
