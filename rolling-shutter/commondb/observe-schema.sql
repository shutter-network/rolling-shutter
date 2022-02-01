CREATE TABLE event_sync_progress (
       id bool UNIQUE NOT NULL DEFAULT true,
       next_block_number integer NOT NULL,
       next_log_index integer NOT NULL
);
INSERT INTO event_sync_progress (next_block_number, next_log_index) VALUES (0,0);

CREATE TABLE keyper_set(
       event_index bigint NOT NULL,
       activation_block_number bigint NOT NULL,
       keypers text[] NOT NULL,
       threshold integer NOT NULL,
       PRIMARY KEY (event_index)
);

CREATE TABLE IF NOT EXISTS decryptor_set_member (
       activation_block_number bigint NOT NULL,
       index int NOT NULL,
       address text NOT NULL,
       PRIMARY KEY (activation_block_number, index)
);

CREATE TABLE chain_collator(
       activation_block_number bigint PRIMARY KEY,
       collator text NOT NULL
);

CREATE TABLE IF NOT EXISTS decryptor_identity (
       address text PRIMARY KEY,
       bls_public_key bytea,
       bls_signature bytea,
       signature_valid boolean NOT NULL
);
