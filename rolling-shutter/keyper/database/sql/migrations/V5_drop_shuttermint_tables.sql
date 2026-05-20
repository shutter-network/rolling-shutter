-- schema-version: keyper-19 --
-- migrations need to start from V2... as file name, as the V1 was initial schema

DROP TABLE IF EXISTS tendermint_batch_config;
DROP TABLE IF EXISTS tendermint_encryption_key;
DROP TABLE IF EXISTS tendermint_outgoing_messages;
DROP TABLE IF EXISTS tendermint_sync_meta;
DROP TABLE IF EXISTS poly_evals;
DROP TABLE IF EXISTS puredkg;
DROP TABLE IF EXISTS outgoing_eon_keys;
DROP TABLE IF EXISTS last_batch_config_sent;
DROP TABLE IF EXISTS last_block_seen;
