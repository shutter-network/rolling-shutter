-- schema-version: 0 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.


CREATE SCHEMA IF NOT EXISTS collator;

CREATE TABLE collator.meta_inf(
       key text PRIMARY KEY,
       value text NOT NULL
);
