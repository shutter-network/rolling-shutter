package metadb

import (
	"context"

	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

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
