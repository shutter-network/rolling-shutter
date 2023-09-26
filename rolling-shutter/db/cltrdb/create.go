// Package cltrdb contains the sqlc generated files for interacting with the collator's database schema.
package cltrdb

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// schemaVersion is used to check that we use the right schema.
var schemaVersion = db.MustFindSchemaVersion("cltrdb")

func initDB(ctx context.Context, tx pgx.Tx) error {
	dbSchemas := []string{
		"cltrdb",
		"chainobsdb/collator",
		"chainobsdb/keyper",
		"chainobsdb/sync",
		"metadb",
	}
	err := db.Create(ctx, tx, dbSchemas)
	if err != nil {
		return err
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
func ValidateDB(ctx context.Context, dbtx DBTX) error {
	return metadb.ValidateSchemaVersion(ctx, dbtx, schemaVersion)
}
