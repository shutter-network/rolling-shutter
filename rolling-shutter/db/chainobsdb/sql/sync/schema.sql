CREATE TABLE event_sync_progress (
       id bool UNIQUE NOT NULL DEFAULT true,
       next_block_number integer NOT NULL,
       next_log_index integer NOT NULL
);
INSERT INTO event_sync_progress (next_block_number, next_log_index) VALUES (0,0);
