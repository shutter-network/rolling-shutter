package snpdb

import (
	"context"
	_ "embed" // blank import

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/metadb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// schemaVersion is used to check that we use the right schema.
var schemaVersion = db.MustFindSchemaVersion("snpdb")

func initSnapshotDB(ctx context.Context, tx pgx.Tx) error {
	err := db.Create(ctx, tx, []string{"snpdb", "chainobsdb", "metadb"})
	if err != nil {
		return err
	}

	err = metadb.New(tx).InsertMeta(
		ctx, metadb.InsertMetaParams{
			Key:   shdb.SchemaVersionKey,
			Value: schemaVersion,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to set schema version in meta_inf table")
	}
	return nil
}

func InitDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return dbpool.BeginFunc(
		ctx, func(tx pgx.Tx) error {
			return initSnapshotDB(ctx, tx)
		},
	)
}

// ValidateSnapshotDB checks that the database schema is compatible.
func ValidateSnapshotDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return metadb.ValidateSchemaVersion(ctx, dbpool, schemaVersion)
}
