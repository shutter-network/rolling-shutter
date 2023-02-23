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

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/commondb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// CreateKeyperTables contains the SQL statements to create the keyper tables.
//
//go:embed schema.sql
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

	_, err = tx.Exec(ctx, metadb.CreateMetaInf)
	if err != nil {
		return errors.Wrap(err, "failed to create meta_inf table")
	}

	err = metadb.New(tx).InsertMeta(ctx, metadb.InsertMetaParams{
		Key: shdb.SchemaVersionKey, Value: schemaVersion,
	})
	if err != nil {
		return errors.Wrap(err, "failed to set schema version in meta_inf table")
	}
	err = New(tx).TMSetSyncMeta(ctx, TMSetSyncMetaParams{
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
func ValidateKeyperDB(ctx context.Context, dbpool DBTX) error {
	return metadb.ValidateSchemaVersion(ctx, dbpool, schemaVersion)
}
