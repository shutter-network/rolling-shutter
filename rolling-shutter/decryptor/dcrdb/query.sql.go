// Code generated by sqlc. DO NOT EDIT.
// source: query.sql

package dcrdb

import (
	"context"
	"database/sql"

	"github.com/jackc/pgconn"
)

const getCipherBatch = `-- name: GetCipherBatch :one
SELECT epoch_id, transactions FROM decryptor.cipher_batch
WHERE epoch_id = $1
`

func (q *Queries) GetCipherBatch(ctx context.Context, epochID []byte) (DecryptorCipherBatch, error) {
	row := q.db.QueryRow(ctx, getCipherBatch, epochID)
	var i DecryptorCipherBatch
	err := row.Scan(&i.EpochID, &i.Transactions)
	return i, err
}

const getDecryptionKey = `-- name: GetDecryptionKey :one
SELECT epoch_id, key FROM decryptor.decryption_key
WHERE epoch_id = $1
`

func (q *Queries) GetDecryptionKey(ctx context.Context, epochID []byte) (DecryptorDecryptionKey, error) {
	row := q.db.QueryRow(ctx, getDecryptionKey, epochID)
	var i DecryptorDecryptionKey
	err := row.Scan(&i.EpochID, &i.Key)
	return i, err
}

const getDecryptionSignature = `-- name: GetDecryptionSignature :one
SELECT epoch_id, signed_hash, signer_index, signature FROM decryptor.decryption_signature
WHERE epoch_id = $1 AND signer_index = $2
`

type GetDecryptionSignatureParams struct {
	EpochID     []byte
	SignerIndex int64
}

func (q *Queries) GetDecryptionSignature(ctx context.Context, arg GetDecryptionSignatureParams) (DecryptorDecryptionSignature, error) {
	row := q.db.QueryRow(ctx, getDecryptionSignature, arg.EpochID, arg.SignerIndex)
	var i DecryptorDecryptionSignature
	err := row.Scan(
		&i.EpochID,
		&i.SignedHash,
		&i.SignerIndex,
		&i.Signature,
	)
	return i, err
}

const getDecryptorIndex = `-- name: GetDecryptorIndex :one
SELECT index
FROM decryptor.decryptor_set_member
WHERE start_epoch_id <= $1 AND address = $2
`

type GetDecryptorIndexParams struct {
	StartEpochID []byte
	Address      sql.NullString
}

func (q *Queries) GetDecryptorIndex(ctx context.Context, arg GetDecryptorIndexParams) (int32, error) {
	row := q.db.QueryRow(ctx, getDecryptorIndex, arg.StartEpochID, arg.Address)
	var index int32
	err := row.Scan(&index)
	return index, err
}

const getDecryptorSet = `-- name: GetDecryptorSet :many
SELECT
    member.start_epoch_id,
    member.index,
    member.address,
    decryptor_identity.bls_public_key
FROM (
    SELECT start_epoch_id, index, address FROM decryptor.decryptor_set_member
    WHERE start_epoch_id <= $1
    ORDER BY start_epoch_id DESC
    FETCH FIRST ROW WITH TIES
) AS member
INNER JOIN decryptor.decryptor_identity
ON member.address=decryptor.decryptor_identity.address
ORDER BY index
`

type GetDecryptorSetRow struct {
	StartEpochID []byte
	Index        int32
	Address      sql.NullString
	BlsPublicKey []byte
}

func (q *Queries) GetDecryptorSet(ctx context.Context, startEpochID []byte) ([]GetDecryptorSetRow, error) {
	rows, err := q.db.Query(ctx, getDecryptorSet, startEpochID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDecryptorSetRow
	for rows.Next() {
		var i GetDecryptorSetRow
		if err := rows.Scan(
			&i.StartEpochID,
			&i.Index,
			&i.Address,
			&i.BlsPublicKey,
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

const getMeta = `-- name: GetMeta :one
SELECT key, value FROM decryptor.meta_inf WHERE key = $1
`

func (q *Queries) GetMeta(ctx context.Context, key string) (DecryptorMetaInf, error) {
	row := q.db.QueryRow(ctx, getMeta, key)
	var i DecryptorMetaInf
	err := row.Scan(&i.Key, &i.Value)
	return i, err
}

const insertCipherBatch = `-- name: InsertCipherBatch :execresult
INSERT INTO decryptor.cipher_batch (
    epoch_id, transactions
) VALUES (
    $1, $2
)
ON CONFLICT DO NOTHING
`

type InsertCipherBatchParams struct {
	EpochID      []byte
	Transactions [][]byte
}

func (q *Queries) InsertCipherBatch(ctx context.Context, arg InsertCipherBatchParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertCipherBatch, arg.EpochID, arg.Transactions)
}

const insertDecryptionKey = `-- name: InsertDecryptionKey :execresult
INSERT INTO decryptor.decryption_key (
    epoch_id, key
) VALUES (
    $1, $2
)
ON CONFLICT DO NOTHING
`

type InsertDecryptionKeyParams struct {
	EpochID []byte
	Key     []byte
}

func (q *Queries) InsertDecryptionKey(ctx context.Context, arg InsertDecryptionKeyParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertDecryptionKey, arg.EpochID, arg.Key)
}

const insertDecryptionSignature = `-- name: InsertDecryptionSignature :execresult
INSERT INTO decryptor.decryption_signature (
    epoch_id, signed_hash, signer_index, signature
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT DO NOTHING
`

type InsertDecryptionSignatureParams struct {
	EpochID     []byte
	SignedHash  []byte
	SignerIndex int64
	Signature   []byte
}

func (q *Queries) InsertDecryptionSignature(ctx context.Context, arg InsertDecryptionSignatureParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertDecryptionSignature,
		arg.EpochID,
		arg.SignedHash,
		arg.SignerIndex,
		arg.Signature,
	)
}

const insertDecryptorIdentity = `-- name: InsertDecryptorIdentity :exec
INSERT INTO decryptor.decryptor_identity (
    address, bls_public_key
) VALUES (
    $1, $2
)
`

type InsertDecryptorIdentityParams struct {
	Address      string
	BlsPublicKey []byte
}

func (q *Queries) InsertDecryptorIdentity(ctx context.Context, arg InsertDecryptorIdentityParams) error {
	_, err := q.db.Exec(ctx, insertDecryptorIdentity, arg.Address, arg.BlsPublicKey)
	return err
}

const insertDecryptorSetMember = `-- name: InsertDecryptorSetMember :exec
INSERT INTO decryptor.decryptor_set_member (
    start_epoch_id, index, address
) VALUES (
    $1, $2, $3
)
`

type InsertDecryptorSetMemberParams struct {
	StartEpochID []byte
	Index        int32
	Address      sql.NullString
}

func (q *Queries) InsertDecryptorSetMember(ctx context.Context, arg InsertDecryptorSetMemberParams) error {
	_, err := q.db.Exec(ctx, insertDecryptorSetMember, arg.StartEpochID, arg.Index, arg.Address)
	return err
}

const insertMeta = `-- name: InsertMeta :exec
INSERT INTO decryptor.meta_inf (key, value) VALUES ($1, $2)
`

type InsertMetaParams struct {
	Key   string
	Value string
}

func (q *Queries) InsertMeta(ctx context.Context, arg InsertMetaParams) error {
	_, err := q.db.Exec(ctx, insertMeta, arg.Key, arg.Value)
	return err
}
