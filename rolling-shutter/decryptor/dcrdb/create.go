// Package dcrdb contains the sqlc generated files for interacting with the decryptor's database
// schema.
package dcrdb

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shuttermint/shdb"
)

//go:embed schema.sql
// CreateDecryptorTables contains the SQL statements to create the decryptor namespace and tables.
var CreateDecryptorTables string

// schemaVersion is used to check that we use the right schema.
var schemaVersion = shdb.MustFindSchemaVersion(CreateDecryptorTables, "dcrdb/schema.sql")

// InitDecryptorDB initializes the database of the decryptor. It is assumed that the db is empty.
func InitDecryptorDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	tx, err := dbpool.Begin(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, CreateDecryptorTables)
	if err != nil {
		return errors.Wrap(err, "failed to create decryptor tables")
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	err = New(dbpool).InsertMeta(ctx, InsertMetaParams{Key: shdb.SchemaVersionKey, Value: schemaVersion})
	if err != nil {
		return errors.Wrap(err, "failed to set schema version in meta_inf table")
	}

	return nil
}

// ValidateDecryptorDB checks that the database schema is compatible.
func ValidateDecryptorDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	m, err := New(dbpool).GetMeta(ctx, shdb.SchemaVersionKey)
	if err != nil {
		return errors.Wrap(err, "failed to get schema version from meta_inf table")
	}
	if m.Value != schemaVersion {
		return errors.Errorf("database has wrong schema version: expected %s, got %s", schemaVersion, m.Value)
	}
	return nil
}
