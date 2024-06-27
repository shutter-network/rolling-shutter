// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: keyper.sql

package database

import (
	"context"
)

const getKeyperSet = `-- name: GetKeyperSet :one
SELECT keyper_config_index, activation_block_number, keypers, threshold FROM keyper_set
WHERE activation_block_number <= $1
ORDER BY activation_block_number DESC LIMIT 1
`

func (q *Queries) GetKeyperSet(ctx context.Context, activationBlockNumber int64) (KeyperSet, error) {
	row := q.db.QueryRow(ctx, getKeyperSet, activationBlockNumber)
	var i KeyperSet
	err := row.Scan(
		&i.KeyperConfigIndex,
		&i.ActivationBlockNumber,
		&i.Keypers,
		&i.Threshold,
	)
	return i, err
}

const getKeyperSetByKeyperConfigIndex = `-- name: GetKeyperSetByKeyperConfigIndex :one
SELECT keyper_config_index, activation_block_number, keypers, threshold FROM keyper_set WHERE keyper_config_index=$1
`

func (q *Queries) GetKeyperSetByKeyperConfigIndex(ctx context.Context, keyperConfigIndex int64) (KeyperSet, error) {
	row := q.db.QueryRow(ctx, getKeyperSetByKeyperConfigIndex, keyperConfigIndex)
	var i KeyperSet
	err := row.Scan(
		&i.KeyperConfigIndex,
		&i.ActivationBlockNumber,
		&i.Keypers,
		&i.Threshold,
	)
	return i, err
}

const getKeyperSets = `-- name: GetKeyperSets :many
SELECT keyper_config_index, activation_block_number, keypers, threshold FROM keyper_set
ORDER BY activation_block_number ASC
`

func (q *Queries) GetKeyperSets(ctx context.Context) ([]KeyperSet, error) {
	rows, err := q.db.Query(ctx, getKeyperSets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []KeyperSet
	for rows.Next() {
		var i KeyperSet
		if err := rows.Scan(
			&i.KeyperConfigIndex,
			&i.ActivationBlockNumber,
			&i.Keypers,
			&i.Threshold,
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

const insertKeyperSet = `-- name: InsertKeyperSet :exec
INSERT INTO keyper_set (
    keyper_config_index,
    activation_block_number,
    keypers,
    threshold
) VALUES (
    $1, $2, $3, $4
) ON CONFLICT DO NOTHING
`

type InsertKeyperSetParams struct {
	KeyperConfigIndex     int64
	ActivationBlockNumber int64
	Keypers               []string
	Threshold             int32
}

func (q *Queries) InsertKeyperSet(ctx context.Context, arg InsertKeyperSetParams) error {
	_, err := q.db.Exec(ctx, insertKeyperSet,
		arg.KeyperConfigIndex,
		arg.ActivationBlockNumber,
		arg.Keypers,
		arg.Threshold,
	)
	return err
}
