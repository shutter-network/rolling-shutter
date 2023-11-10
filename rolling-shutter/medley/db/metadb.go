package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var (
	ErrValueMismatch = errors.New("database has unexpected value")
	ErrKeyNotFound   = errors.New("key does not exist")
)

var DatabaseVersionKey string = "database-version"

// MakeSchemaVersionKey generates the version key for the metadb
// that is used to check database schema compatibility.
// `definitionName` is the name of the database definition that corresponds
// to a database subset (one entry in an sqlc file).
// `schemaName` corresponds to the individual *.sql files in the schema folder.
func MakeSchemaVersionKey(definitionName, schemaName string) string {
	return "schema-version-" + definitionName + "-" + schemaName
}

func InsertDBVersion(ctx context.Context, tx pgx.Tx, version string) error {
	return insertMetaInf(ctx, tx, DatabaseVersionKey, version)
}

func InsertSchemaVersion(ctx context.Context, tx pgx.Tx, definitionName string, schema Schema) error {
	return insertMetaInf(ctx, tx, MakeSchemaVersionKey(definitionName, schema.Name), fmt.Sprint(schema.Version))
}

func insertMetaInf(ctx context.Context, tx pgx.Tx, key, val string) error {
	log.Info().Str("key", key).
		Str("value", val).Msg("insert schema meta inf")
	return New(tx).InsertMeta(ctx, InsertMetaParams{
		Key:   key,
		Value: val,
	})
}

// ValidateSchemaVersion checks that the database schema is compatible.
func ValidateSchemaVersion(ctx context.Context, tx pgx.Tx, definitionName string, schema Schema) error {
	return expectMetaKeyVal(ctx, tx, MakeSchemaVersionKey(definitionName, schema.Name), fmt.Sprint(schema.Version))
}

func expectMetaKeyVal(ctx context.Context, tx pgx.Tx, key, val string) error {
	haveVal, err := New(tx).GetMeta(ctx, key)
	if err == pgx.ErrNoRows {
		return errors.Wrapf(ErrKeyNotFound, "key: %s", key)
	} else if err != nil {
		return errors.Wrapf(err, "failed to get key '%s' from meta_inf table", key)
	}
	if haveVal != val {
		return errors.Wrapf(ErrValueMismatch, "expected %s, have %s", val, haveVal)
	}
	return nil
}

// ValidateDatabaseVersion checks that the overall database is compatible.
// This corresponds to the "role" of the database, e.g. a snapshot-keyper
// might not be compatible with a snapshot test-keyper, even if the schema's
// versions would match exactly.
func ValidateDatabaseVersion(ctx context.Context, tx pgx.Tx, version string) error {
	return expectMetaKeyVal(ctx, tx, DatabaseVersionKey, version)
}
