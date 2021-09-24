// Package kprdb contains the sqlc generated files for interacting with the keyper's database
// schema.
package kprdb

import (
	"context"
	_ "embed"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shuttermint/shdb"
)

//go:embed schema.sql
// CreateKeyperTables contains the SQL statements to create the keyper namespace and tables.
var CreateKeyperTables string

// schemaVersion is used to check that we use the right schema.
var schemaVersion = shdb.MustFindSchemaVersion(CreateKeyperTables, "kprdb/schema.sql")

func initKeyperDB(ctx context.Context, tx pgx.Tx, queries *Queries) error {
	_, err := tx.Exec(ctx, CreateKeyperTables)
	if err != nil {
		return errors.Wrap(err, "failed to create keyper tables")
	}

	err = queries.InsertMeta(ctx, InsertMetaParams{Key: shdb.SchemaVersionKey, Value: schemaVersion})
	if err != nil {
		return errors.Wrap(err, "failed to set schema version in meta_inf table")
	}
	err = queries.TMSetSyncMeta(ctx, TMSetSyncMetaParams{
		CurrentBlock:        -1,
		LastCommittedHeight: -1,
		SyncTimestamp:       time.Now(),
	})
	if err != nil {
		return errors.Wrap(err, "failed to set current block")
	}
	return nil
}

// InitKeyperDB initializes the database of the keyper. It is assumed that the db is empty.
func InitKeyperDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	tx, err := dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to start tx")
	}
	queries := New(dbpool).WithTx(tx)
	err = initKeyperDB(ctx, tx, queries)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// ValidateKeyperDB checks that the database schema is compatible.
func ValidateKeyperDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	m, err := New(dbpool).GetMeta(ctx, shdb.SchemaVersionKey)
	if err != nil {
		return errors.Wrap(err, "failed to get schema version from meta_inf table")
	}
	if m.Value != schemaVersion {
		return errors.Errorf("database has wrong schema version: expected %s, got %s", schemaVersion, m.Value)
	}
	return nil
}
