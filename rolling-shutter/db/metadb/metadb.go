package metadb

import (
	"context"
	_ "embed"

	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// CreateMetaInf contains the SQL statements to create the meta_inf table.
//
//go:embed schema.sql
var CreateMetaInf string

// ValidateSchemaVersion checks that the database schema is compatible.
func ValidateSchemaVersion(ctx context.Context, db DBTX, expectedSchemaVersion string) error {
	val, err := New(db).GetMeta(ctx, shdb.SchemaVersionKey)
	if err != nil {
		return errors.Wrap(err, "failed to get schema version from meta_inf table")
	}
	if val != expectedSchemaVersion {
		return errors.Errorf("database has wrong schema version: expected %s, got %s",
			expectedSchemaVersion,
			val,
		)
	}
	return nil
}
