-- name: InsertDecryptionKey :execresult
INSERT INTO decryption_key (eon, epoch_id, decryption_key)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetDecryptionKey :one
SELECT * FROM decryption_key
WHERE eon = $1 AND epoch_id = $2;

-- name: ExistsDecryptionKey :one
SELECT EXISTS (
    SELECT 1
    FROM decryption_key
    WHERE eon = $1 AND epoch_id = $2
);

-- name: InsertDecryptionKeyShare :exec
INSERT INTO decryption_key_share (eon, epoch_id, keyper_index, decryption_key_share)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING;

-- name: SelectDecryptionKeyShares :many
SELECT * FROM decryption_key_share
WHERE eon = $1 AND epoch_id = $2;

-- name: GetDecryptionKeyShare :one
SELECT * FROM decryption_key_share
WHERE eon = $1 AND epoch_id = $2 AND keyper_index = $3;

-- name: ExistsDecryptionKeyShare :one
SELECT EXISTS (
    SELECT 1
    FROM decryption_key_share
    WHERE eon = $1 AND epoch_id = $2 AND keyper_index = $3
);

-- name: CountDecryptionKeyShares :one
SELECT count(*) FROM decryption_key_share
WHERE eon = $1 AND epoch_id = $2;

-- name: InsertEon :exec
INSERT INTO eons (eon, activation_block_number, keyper_config_index, dkg_contract, phase_length, lead_length, max_retries)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetEon :one
SELECT * FROM eons WHERE eon=$1;

-- name: GetEonForBlockNumber :one
SELECT * FROM eons
WHERE activation_block_number <= sqlc.arg(block_number)
ORDER BY activation_block_number DESC
LIMIT 1;

-- name: GetAllEons :many
SELECT * FROM eons ORDER BY eon;

-- name: InsertDKGResult :exec
INSERT INTO dkg_result (eon,success,error,pure_result)
VALUES ($1,$2,$3,$4);

-- name: GetDKGResult :one
SELECT * FROM dkg_result
WHERE eon = $1;

-- name: ExistsDKGResultSuccess :one
SELECT EXISTS (
    SELECT 1 FROM dkg_result WHERE eon = $1 AND success = TRUE
);

-- name: GetDKGResultForBlockNumber :one
SELECT * FROM dkg_result
WHERE eon = (SELECT eon FROM eons WHERE activation_block_number <= sqlc.arg(block_number)
ORDER BY activation_block_number DESC
LIMIT 1);

-- name: GetDKGResultForKeyperConfigIndex :one
SELECT * FROM dkg_result
WHERE eon = (SELECT max(eon) FROM eons WHERE keyper_config_index = $1);

-- name: GetAllDKGResults :many
SELECT * FROM dkg_result
ORDER BY eon ASC;

-- name: GetLatestEonForKeyperConfig :one
SELECT max(eons.eon)::INT
FROM eons
WHERE eons.keyper_config_index = @keyper_config_index;

-- name: GetLatestStartedEonByKeyperConfigIndex :one
SELECT *
FROM eons
WHERE keyper_config_index = $1
ORDER BY eon DESC
LIMIT 1;

-- name: UpsertECIESKey :exec
INSERT INTO ecies_keys (keyper_address, ecies_public_key)
VALUES ($1, $2)
ON CONFLICT (keyper_address) DO UPDATE
SET ecies_public_key = EXCLUDED.ecies_public_key;

-- name: GetECIESKey :one
SELECT * FROM ecies_keys WHERE keyper_address = $1;

-- name: ExistsECIESKey :one
SELECT EXISTS (
    SELECT 1 FROM ecies_keys WHERE keyper_address = $1
);

-- name: InsertDKGPolyCommitment :exec
INSERT INTO dkg_poly_commitments (keyper_set_index, retry_counter, keyper_index, commitment)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING;

-- name: GetDKGPolyCommitments :many
SELECT * FROM dkg_poly_commitments
WHERE keyper_set_index = $1 AND retry_counter = $2
ORDER BY keyper_index;

-- name: InsertDKGPolyEval :exec
INSERT INTO dkg_poly_evals (keyper_set_index, retry_counter, sender_index, receiver_index, encrypted_eval)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING;

-- name: GetDKGPolyEvals :many
SELECT * FROM dkg_poly_evals
WHERE keyper_set_index = $1 AND retry_counter = $2
ORDER BY sender_index, receiver_index;

-- name: InsertDKGAccusation :exec
INSERT INTO dkg_accusations (keyper_set_index, retry_counter, accuser_index, accused_index)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING;

-- name: GetDKGAccusations :many
SELECT * FROM dkg_accusations
WHERE keyper_set_index = $1 AND retry_counter = $2
ORDER BY accuser_index, accused_index;

-- name: InsertDKGApology :exec
INSERT INTO dkg_apologies (keyper_set_index, retry_counter, apologizer_index, accuser_index, poly_eval)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO NOTHING;

-- name: GetDKGApologies :many
SELECT * FROM dkg_apologies
WHERE keyper_set_index = $1 AND retry_counter = $2
ORDER BY apologizer_index, accuser_index;

-- name: InsertDKGInitialState :exec
INSERT INTO dkg_initial_states (keyper_set_index, retry_counter, puredkg_bytes)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetDKGInitialState :one
SELECT * FROM dkg_initial_states
WHERE keyper_set_index = $1 AND retry_counter = $2;

-- name: InsertDKGSentAction :exec
INSERT INTO dkg_sent_actions (keyper_set_index, retry_counter, action, tx_outbox_id)
VALUES ($1, $2, $3, $4);

-- name: ExistsDKGSentAction :one
SELECT EXISTS (
    SELECT 1 FROM dkg_sent_actions
    WHERE keyper_set_index = $1 AND retry_counter = $2 AND action = $3
);

-- name: InsertPendingTx :one
INSERT INTO tx_outbox (to_address, data, value, label)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetPendingTxs :many
SELECT * FROM tx_outbox
WHERE status = 'pending'
ORDER BY id;

-- name: GetSubmittedTxs :many
SELECT * FROM tx_outbox
WHERE status = 'submitted'
ORDER BY id;

-- name: GetTxOutboxByID :one
SELECT * FROM tx_outbox WHERE id = $1;

-- name: MarkTxSubmitted :exec
UPDATE tx_outbox
SET status = 'submitted', tx_hash = $2, nonce = $3, updated_at = NOW()
WHERE id = $1;

-- name: MarkTxConfirmed :exec
UPDATE tx_outbox
SET status = 'confirmed', updated_at = NOW()
WHERE id = $1;

-- name: MarkTxFailed :exec
UPDATE tx_outbox
SET status = 'failed', error = $2, updated_at = NOW()
WHERE id = $1;
