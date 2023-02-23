// Package cltrdb contains the sqlc generated files for interacting with the collator's database schema.
package cltrdb

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/commondb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// CreateCollatorTables contains the SQL statements to create the collator tables.
//
//go:embed schema.sql
var CreateCollatorTables string

// schemaVersion is used to check that we use the right schema.
var schemaVersion = shdb.MustFindSchemaVersion(CreateCollatorTables, "cltrdb/schema.sql")

func initDB(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, CreateCollatorTables)
	if err != nil {
		return errors.Wrap(err, "failed to create collator tables")
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
		Key:   shdb.SchemaVersionKey,
		Value: schemaVersion,
	})
	if err != nil {
		return errors.Wrap(err, "failed to set schema version in meta_inf table")
	}
	return nil
}

// InitDB initializes the database of the collator. It is assumed that the db is empty.
func InitDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return initDB(ctx, tx)
	})
}

// ValidateDB checks that the database schema is compatible.
func ValidateDB(ctx context.Context, db DBTX) error {
	return metadb.ValidateSchemaVersion(ctx, db, schemaVersion)
}
