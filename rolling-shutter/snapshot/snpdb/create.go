package snpdb

import (
	"context"
	_ "embed"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/shutter-network/shutter/shuttermint/commondb"
	"github.com/shutter-network/shutter/shuttermint/shdb"

	"github.com/jackc/pgx/v4/pgxpool"
)

//go:embed schema.sql
// CreateSnapshotTables contains the SQL statements to create the decryptor tables.
var CreateSnapshotTables string

// schemaVersion is used to check that we use the right schema.
var schemaVersion = shdb.MustFindSchemaVersion(CreateSnapshotTables, "snpdb/schema.sql")

func initSnapshotDB(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, CreateSnapshotTables)
	if err != nil {
		return errors.Wrap(err, "failed to create snapshot tables")
	}

	// FIXME: Add back when / if necessary
	//_, err = tx.Exec(ctx, commondb.CreateObserveTables)
	//if err != nil {
	//	return errors.Wrap(err, "failed to create observe tables")
	//}

	_, err = tx.Exec(ctx, commondb.CreateMetaInf)
	if err != nil {
		return errors.Wrap(err, "failed to create meta_inf table")
	}
	err = New(tx).InsertMeta(ctx, InsertMetaParams{Key: shdb.SchemaVersionKey, Value: schemaVersion})
	if err != nil {
		return errors.Wrap(err, "failed to set schema version in meta_inf table")
	}

	return nil
}

func InitDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return initSnapshotDB(ctx, tx)
	})
}

// ValidateSnapshotDB checks that the database schema is compatible.
func ValidateSnapshotDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	m, err := New(dbpool).GetMeta(ctx, shdb.SchemaVersionKey)
	if err != nil {
		return errors.Wrap(err, "failed to get schema version from meta_inf table")
	}
	if m != schemaVersion {
		return errors.Errorf("database has wrong schema version: expected %s, got %s", schemaVersion, m)
	}
	return nil
}
