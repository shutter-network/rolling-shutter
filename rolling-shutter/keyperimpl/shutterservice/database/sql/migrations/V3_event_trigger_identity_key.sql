ALTER TABLE fired_triggers
    ADD COLUMN IF NOT EXISTS identity bytea;

UPDATE fired_triggers f
SET identity = e.identity
    FROM event_trigger_registered_event e
WHERE f.eon = e.eon
  AND f.identity_prefix = e.identity_prefix
  AND f.sender = e.sender;

ALTER TABLE fired_triggers
    ALTER COLUMN identity SET NOT NULL;

ALTER TABLE fired_triggers
DROP CONSTRAINT IF EXISTS fired_triggers_eon_identity_prefix_sender_fkey;

ALTER TABLE fired_triggers
DROP CONSTRAINT IF EXISTS fired_triggers_pkey;

ALTER TABLE event_trigger_registered_event
DROP CONSTRAINT IF EXISTS event_trigger_registered_event_pkey;

ALTER TABLE event_trigger_registered_event
    ADD CONSTRAINT event_trigger_registered_event_pkey PRIMARY KEY (eon, identity);

ALTER TABLE fired_triggers
    ADD CONSTRAINT fired_triggers_pkey PRIMARY KEY (eon, identity);

ALTER TABLE fired_triggers
    ADD CONSTRAINT fired_triggers_eon_identity_fkey
        FOREIGN KEY (eon, identity)
            REFERENCES event_trigger_registered_event (eon, identity)
            ON DELETE CASCADE;
