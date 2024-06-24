CREATE TABLE keyper_set(
       keyper_config_index bigint NOT NULL,
       activation_block_number bigint NOT NULL,
       keypers text[] NOT NULL,
       threshold integer NOT NULL,
       PRIMARY KEY (keyper_config_index)
);

CREATE TABLE recent_block (
       block_hash bytea NOT NULL,
       block_number bigint NOT NULL,
       parent_hash bytea NOT NULL,
       timestamp bigint NOT NULL,
       header bytea NOT NULL,
       PRIMARY KEY (block_number)
);