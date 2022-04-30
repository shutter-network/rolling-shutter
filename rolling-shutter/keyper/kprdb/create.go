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

	"github.com/shutter-network/rolling-shutter/rolling-shutter/commondb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

//go:embed schema.sql
// CreateKeyperTables contains the SQL statements to create the keyper tables.
var CreateKeyperTables string

// schemaVersion is used to check that we use the right schema.
var schemaVersion = shdb.MustFindSchemaVersion(CreateKeyperTables, "kprdb/schema.sql")

func initDB(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, CreateKeyperTables)
	if err != nil {
		return errors.Wrap(err, "failed to create keyper tables")
	}
	_, err = tx.Exec(ctx, commondb.CreateObserveTables)
	if err != nil {
		return errors.Wrap(err, "failed to create observe tables")
	}

	_, err = tx.Exec(ctx, commondb.CreateMetaInf)
	if err != nil {
		return errors.Wrap(err, "failed to create meta_inf table")
	}

	queries := New(tx)
	err = queries.InsertMeta(ctx, InsertMetaParams{Key: shdb.SchemaVersionKey, Value: schemaVersion})
	if err != nil {
		return errors.Wrap(err, "failed to set schema version in meta_inf table")
	}
	err = queries.TMSetSyncMeta(ctx, TMSetSyncMetaParams{
		CurrentBlock:        0,
		LastCommittedHeight: -1,
		SyncTimestamp:       time.Now(),
	})
	if err != nil {
		return errors.Wrap(err, "failed to set current block")
	}
	return nil
}

// InitDB initializes the database of the keyper. It is assumed that the db is empty.
func InitDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return initDB(ctx, tx)
	})
}

// ValidateKeyperDB checks that the database schema is compatible.
func ValidateKeyperDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	val, err := New(dbpool).GetMeta(ctx, shdb.SchemaVersionKey)
	if err != nil {
		return errors.Wrap(err, "failed to get schema version from meta_inf table")
	}
	if val != schemaVersion {
		return errors.Errorf("database has wrong schema version: expected %s, got %s", schemaVersion, val)
	}
	return nil
}
