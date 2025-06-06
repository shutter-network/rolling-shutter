// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: shutterservice.sql

package database

import (
	"context"

	"github.com/jackc/pgconn"
)

const deleteIdentityRegisteredEventsFromBlockNumber = `-- name: DeleteIdentityRegisteredEventsFromBlockNumber :exec
DELETE FROM identity_registered_event WHERE block_number >= $1
`

func (q *Queries) DeleteIdentityRegisteredEventsFromBlockNumber(ctx context.Context, blockNumber int64) error {
	_, err := q.db.Exec(ctx, deleteIdentityRegisteredEventsFromBlockNumber, blockNumber)
	return err
}

const getCurrentDecryptionTrigger = `-- name: GetCurrentDecryptionTrigger :one
SELECT eon, triggered_block_number, identities_hash FROM current_decryption_trigger
WHERE eon = $1 ORDER BY triggered_block_number DESC LIMIT 1
`

func (q *Queries) GetCurrentDecryptionTrigger(ctx context.Context, eon int64) (CurrentDecryptionTrigger, error) {
	row := q.db.QueryRow(ctx, getCurrentDecryptionTrigger, eon)
	var i CurrentDecryptionTrigger
	err := row.Scan(&i.Eon, &i.TriggeredBlockNumber, &i.IdentitiesHash)
	return i, err
}

const getDecryptionSignatures = `-- name: GetDecryptionSignatures :many
SELECT eon, keyper_index, identities_hash, signature FROM decryption_signatures
WHERE eon = $1 AND identities_hash = $2
ORDER BY keyper_index ASC
LIMIT $3
`

type GetDecryptionSignaturesParams struct {
	Eon            int64
	IdentitiesHash []byte
	Limit          int32
}

func (q *Queries) GetDecryptionSignatures(ctx context.Context, arg GetDecryptionSignaturesParams) ([]DecryptionSignature, error) {
	rows, err := q.db.Query(ctx, getDecryptionSignatures, arg.Eon, arg.IdentitiesHash, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DecryptionSignature
	for rows.Next() {
		var i DecryptionSignature
		if err := rows.Scan(
			&i.Eon,
			&i.KeyperIndex,
			&i.IdentitiesHash,
			&i.Signature,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getIdentityRegisteredEventsSyncedUntil = `-- name: GetIdentityRegisteredEventsSyncedUntil :one
SELECT enforce_one_row, block_hash, block_number FROM identity_registered_events_synced_until LIMIT 1
`

func (q *Queries) GetIdentityRegisteredEventsSyncedUntil(ctx context.Context) (IdentityRegisteredEventsSyncedUntil, error) {
	row := q.db.QueryRow(ctx, getIdentityRegisteredEventsSyncedUntil)
	var i IdentityRegisteredEventsSyncedUntil
	err := row.Scan(&i.EnforceOneRow, &i.BlockHash, &i.BlockNumber)
	return i, err
}

const getNotDecryptedIdentityRegisteredEvents = `-- name: GetNotDecryptedIdentityRegisteredEvents :many
SELECT block_number, block_hash, tx_index, log_index, eon, identity_prefix, sender, timestamp, decrypted, identity FROM identity_registered_event
WHERE timestamp >= $1 AND timestamp <= $2 AND decrypted = false
ORDER BY timestamp ASC
`

type GetNotDecryptedIdentityRegisteredEventsParams struct {
	Timestamp   int64
	Timestamp_2 int64
}

func (q *Queries) GetNotDecryptedIdentityRegisteredEvents(ctx context.Context, arg GetNotDecryptedIdentityRegisteredEventsParams) ([]IdentityRegisteredEvent, error) {
	rows, err := q.db.Query(ctx, getNotDecryptedIdentityRegisteredEvents, arg.Timestamp, arg.Timestamp_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []IdentityRegisteredEvent
	for rows.Next() {
		var i IdentityRegisteredEvent
		if err := rows.Scan(
			&i.BlockNumber,
			&i.BlockHash,
			&i.TxIndex,
			&i.LogIndex,
			&i.Eon,
			&i.IdentityPrefix,
			&i.Sender,
			&i.Timestamp,
			&i.Decrypted,
			&i.Identity,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertDecryptionSignature = `-- name: InsertDecryptionSignature :exec
INSERT INTO decryption_signatures (eon, keyper_index, identities_hash, signature)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING
`

type InsertDecryptionSignatureParams struct {
	Eon            int64
	KeyperIndex    int64
	IdentitiesHash []byte
	Signature      []byte
}

func (q *Queries) InsertDecryptionSignature(ctx context.Context, arg InsertDecryptionSignatureParams) error {
	_, err := q.db.Exec(ctx, insertDecryptionSignature,
		arg.Eon,
		arg.KeyperIndex,
		arg.IdentitiesHash,
		arg.Signature,
	)
	return err
}

const insertIdentityRegisteredEvent = `-- name: InsertIdentityRegisteredEvent :execresult
INSERT INTO identity_registered_event (
    block_number,
    block_hash,
    tx_index,
    log_index,
    eon,
    identity_prefix,
    sender,
    timestamp,
    identity
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (identity_prefix, sender) DO UPDATE SET
block_number = $1,
block_hash = $2,
tx_index = $3,
log_index = $4,
sender = $7,
timestamp = $8,
identity = $9
`

type InsertIdentityRegisteredEventParams struct {
	BlockNumber    int64
	BlockHash      []byte
	TxIndex        int64
	LogIndex       int64
	Eon            int64
	IdentityPrefix []byte
	Sender         string
	Timestamp      int64
	Identity       []byte
}

func (q *Queries) InsertIdentityRegisteredEvent(ctx context.Context, arg InsertIdentityRegisteredEventParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertIdentityRegisteredEvent,
		arg.BlockNumber,
		arg.BlockHash,
		arg.TxIndex,
		arg.LogIndex,
		arg.Eon,
		arg.IdentityPrefix,
		arg.Sender,
		arg.Timestamp,
		arg.Identity,
	)
}

const setCurrentDecryptionTrigger = `-- name: SetCurrentDecryptionTrigger :exec
INSERT INTO current_decryption_trigger (eon, triggered_block_number, identities_hash)
VALUES ($1, $2, $3)
ON CONFLICT (eon, triggered_block_number) DO UPDATE
SET triggered_block_number = $2, identities_hash = $3
`

type SetCurrentDecryptionTriggerParams struct {
	Eon                  int64
	TriggeredBlockNumber int64
	IdentitiesHash       []byte
}

func (q *Queries) SetCurrentDecryptionTrigger(ctx context.Context, arg SetCurrentDecryptionTriggerParams) error {
	_, err := q.db.Exec(ctx, setCurrentDecryptionTrigger, arg.Eon, arg.TriggeredBlockNumber, arg.IdentitiesHash)
	return err
}

const setIdentityRegisteredEventSyncedUntil = `-- name: SetIdentityRegisteredEventSyncedUntil :exec
INSERT INTO identity_registered_events_synced_until (block_hash, block_number) VALUES ($1, $2)
ON CONFLICT (enforce_one_row) DO UPDATE
SET block_hash = $1, block_number = $2
`

type SetIdentityRegisteredEventSyncedUntilParams struct {
	BlockHash   []byte
	BlockNumber int64
}

func (q *Queries) SetIdentityRegisteredEventSyncedUntil(ctx context.Context, arg SetIdentityRegisteredEventSyncedUntilParams) error {
	_, err := q.db.Exec(ctx, setIdentityRegisteredEventSyncedUntil, arg.BlockHash, arg.BlockNumber)
	return err
}

const updateDecryptedFlag = `-- name: UpdateDecryptedFlag :exec
UPDATE identity_registered_event
SET decrypted = TRUE
WHERE (eon, identity) IN (
    SELECT UNNEST($1::bigint[]), UNNEST($2::bytea[])
)
`

type UpdateDecryptedFlagParams struct {
	Column1 []int64
	Column2 [][]byte
}

func (q *Queries) UpdateDecryptedFlag(ctx context.Context, arg UpdateDecryptedFlagParams) error {
	_, err := q.db.Exec(ctx, updateDecryptedFlag, arg.Column1, arg.Column2)
	return err
}
