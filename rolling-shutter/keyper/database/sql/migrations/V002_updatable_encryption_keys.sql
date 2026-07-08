ALTER TABLE tendermint_encryption_key
    ADD COLUMN height bigint
        NOT NULL
        DEFAULT 0;

ALTER TABLE tendermint_encryption_key
    DROP CONSTRAINT tendermint_encryption_key_pkey;

ALTER TABLE tendermint_encryption_key
    ADD PRIMARY KEY (address, height);
