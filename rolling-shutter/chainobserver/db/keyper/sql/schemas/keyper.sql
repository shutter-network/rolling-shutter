CREATE TABLE keyper_set(
       keyper_config_index bigint NOT NULL,
       activation_block_number bigint NOT NULL,
       keypers text[] NOT NULL,
       threshold integer NOT NULL,
       PRIMARY KEY (keyper_config_index)
);
