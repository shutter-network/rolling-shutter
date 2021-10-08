// Code generated by sqlc. DO NOT EDIT.
// source: query.sql

package dcrdb

import (
	"context"

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
SELECT epoch_id, signed_hash, signers_bitfield, signature FROM decryptor.decryption_signature
WHERE epoch_id = $1 AND signers_bitfield = $2
`

type GetDecryptionSignatureParams struct {
	EpochID         []byte
	SignersBitfield []byte
}

func (q *Queries) GetDecryptionSignature(ctx context.Context, arg GetDecryptionSignatureParams) (DecryptorDecryptionSignature, error) {
	row := q.db.QueryRow(ctx, getDecryptionSignature, arg.EpochID, arg.SignersBitfield)
	var i DecryptorDecryptionSignature
	err := row.Scan(
		&i.EpochID,
		&i.SignedHash,
		&i.SignersBitfield,
		&i.Signature,
	)
	return i, err
}

const getDecryptionSignatures = `-- name: GetDecryptionSignatures :many
SELECT epoch_id, signed_hash, signers_bitfield, signature FROM decryptor.decryption_signature
WHERE epoch_id = $1 AND signed_hash = $2
`

type GetDecryptionSignaturesParams struct {
	EpochID    []byte
	SignedHash []byte
}

func (q *Queries) GetDecryptionSignatures(ctx context.Context, arg GetDecryptionSignaturesParams) ([]DecryptorDecryptionSignature, error) {
	rows, err := q.db.Query(ctx, getDecryptionSignatures, arg.EpochID, arg.SignedHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DecryptorDecryptionSignature
	for rows.Next() {
		var i DecryptorDecryptionSignature
		if err := rows.Scan(
			&i.EpochID,
			&i.SignedHash,
			&i.SignersBitfield,
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

const getDecryptorIndex = `-- name: GetDecryptorIndex :one
SELECT index
FROM decryptor.decryptor_set_member
WHERE start_epoch_id <= $1 AND address = $2
`

type GetDecryptorIndexParams struct {
	StartEpochID []byte
	Address      string
}

func (q *Queries) GetDecryptorIndex(ctx context.Context, arg GetDecryptorIndexParams) (int32, error) {
	row := q.db.QueryRow(ctx, getDecryptorIndex, arg.StartEpochID, arg.Address)
	var index int32
	err := row.Scan(&index)
	return index, err
}

const getDecryptorKey = `-- name: GetDecryptorKey :one
SELECT bls_public_key FROM decryptor.decryptor_identity WHERE address = (
    SELECT address from decryptor.decryptor_set_member
    WHERE index = $1 AND start_epoch_id <= $2 ORDER BY start_epoch_id DESC LIMIT 1
)
`

type GetDecryptorKeyParams struct {
	Index        int32
	StartEpochID []byte
}

func (q *Queries) GetDecryptorKey(ctx context.Context, arg GetDecryptorKeyParams) ([]byte, error) {
	row := q.db.QueryRow(ctx, getDecryptorKey, arg.Index, arg.StartEpochID)
	var bls_public_key []byte
	err := row.Scan(&bls_public_key)
	return bls_public_key, err
}

const getDecryptorSet = `-- name: GetDecryptorSet :many
SELECT
    member.start_epoch_id,
    member.index,
    member.address,
    identity.bls_public_key
FROM (
    SELECT
        start_epoch_id,
        index,
        address
    FROM decryptor.decryptor_set_member
    WHERE start_epoch_id = (
        SELECT
            m.start_epoch_id
        FROM decryptor.decryptor_set_member AS m
        WHERE m.start_epoch_id <= $1
        ORDER BY m.start_epoch_id DESC
        LIMIT 1
    )
) AS member
LEFT OUTER JOIN decryptor.decryptor_identity AS identity
ON member.address = identity.address
ORDER BY index
`

type GetDecryptorSetRow struct {
	StartEpochID []byte
	Index        int32
	Address      string
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

const getEonPublicKey = `-- name: GetEonPublicKey :one
SELECT eon_public_key
FROM decryptor.eon_public_key
WHERE start_epoch_id <= $1
ORDER BY start_epoch_id DESC
LIMIT 1
`

func (q *Queries) GetEonPublicKey(ctx context.Context, startEpochID []byte) ([]byte, error) {
	row := q.db.QueryRow(ctx, getEonPublicKey, startEpochID)
	var eon_public_key []byte
	err := row.Scan(&eon_public_key)
	return eon_public_key, err
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
    epoch_id, signed_hash, signers_bitfield, signature
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT DO NOTHING
`

type InsertDecryptionSignatureParams struct {
	EpochID         []byte
	SignedHash      []byte
	SignersBitfield []byte
	Signature       []byte
}

func (q *Queries) InsertDecryptionSignature(ctx context.Context, arg InsertDecryptionSignatureParams) (pgconn.CommandTag, error) {
	return q.db.Exec(ctx, insertDecryptionSignature,
		arg.EpochID,
		arg.SignedHash,
		arg.SignersBitfield,
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
	Address      string
}

func (q *Queries) InsertDecryptorSetMember(ctx context.Context, arg InsertDecryptorSetMemberParams) error {
	_, err := q.db.Exec(ctx, insertDecryptorSetMember, arg.StartEpochID, arg.Index, arg.Address)
	return err
}

const insertEonPublicKey = `-- name: InsertEonPublicKey :exec
INSERT INTO decryptor.eon_public_key (
    start_epoch_id,
    eon_public_key
) VALUES (
    $1, $2
)
`

type InsertEonPublicKeyParams struct {
	StartEpochID []byte
	EonPublicKey []byte
}

func (q *Queries) InsertEonPublicKey(ctx context.Context, arg InsertEonPublicKeyParams) error {
	_, err := q.db.Exec(ctx, insertEonPublicKey, arg.StartEpochID, arg.EonPublicKey)
	return err
}

const insertKeyperSet = `-- name: InsertKeyperSet :exec
INSERT INTO decryptor.keyper_set (
    start_epoch_id,
    keypers,
    threshold
) VALUES (
    $1, $2, $3
)
`

type InsertKeyperSetParams struct {
	StartEpochID []byte
	Keypers      []string
	Threshold    int32
}

func (q *Queries) InsertKeyperSet(ctx context.Context, arg InsertKeyperSetParams) error {
	_, err := q.db.Exec(ctx, insertKeyperSet, arg.StartEpochID, arg.Keypers, arg.Threshold)
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
