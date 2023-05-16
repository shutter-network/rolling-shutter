package metadb

import (
	"context"

	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

var ErrSchemaMismatch = errors.New("database has wrong schema version")

// ValidateSchemaVersion checks that the database schema is compatible.
func ValidateSchemaVersion(ctx context.Context, db DBTX, expectedSchemaVersion string) error {
	val, err := New(db).GetMeta(ctx, shdb.SchemaVersionKey)
	if err != nil {
		return errors.Wrap(err, "failed to get schema version from meta_inf table")
	}
	if val != expectedSchemaVersion {
		return errors.Wrapf(ErrSchemaMismatch, "expected %s, have %s", expectedSchemaVersion, val)
	}
	return nil
}
