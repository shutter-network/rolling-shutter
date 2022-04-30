// Package cltrdb contains the sqlc generated files for interacting with the collator's database schema.
package cltrdb

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/commondb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

//go:embed schema.sql
// CreateCollatorTables contains the SQL statements to create the collator tables.
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
	_, err = tx.Exec(ctx, commondb.CreateMetaInf)
	if err != nil {
		return errors.Wrap(err, "failed to create meta_inf table")
	}
	err = New(tx).InsertMeta(ctx, InsertMetaParams{
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
func ValidateDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	m, err := New(dbpool).GetMeta(ctx, shdb.SchemaVersionKey)
	if err != nil {
		return errors.Wrap(err, "failed to get schema version from meta_inf table")
	}
	if m != schemaVersion {
		return errors.Errorf("database has wrong schema version: expected %s, got %s", schemaVersion, m)
	}
	return nil
}
