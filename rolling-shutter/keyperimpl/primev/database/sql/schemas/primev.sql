-- schema-version: primev-1 --
-- Please change the version above if you make incompatible changes to
-- the schema. We'll use this to check we're using the right schema.

CREATE TABLE commitment(
    tx_hashes text[] NOT NULL,
    provider_address text NOT NULL,
    commitment_signature text NOT NULL,
    commitment_digest text NOT NULL,
    block_number bigint NOT NULL CHECK (block_number >= 0),
    PRIMARY KEY (provider_address, commitment_digest, block_number)
);

CREATE TABLE committed_transactions(
    eon bigint NOT NULL CHECK (eon >= 0),
    identity_preimage text NOT NULL,
    block_number bigint NOT NULL CHECK (block_number >= 0),
    tx_hash text NOT NULL,
    PRIMARY KEY (eon, identity_preimage, tx_hash)
);