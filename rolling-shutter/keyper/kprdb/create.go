package kprdb

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shuttermint/shdb"
)

//go:embed schema.sql
// CreateKeyperTables contains the SQL statements to create the keyper namespace and tables.
var CreateKeyperTables string

// InitKeyperDB initializes the database of the keyper. It is assumed that the db is empty.
func InitKeyperDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	_, err := dbpool.Exec(ctx, CreateKeyperTables)
	if err != nil {
		return errors.Wrap(err, "failed to create keyper tables")
	}
	return nil
}

// ValidateKeyperDB checks that all expected tables exist in the database. If not, it returns an
// error.
func ValidateKeyperDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return shdb.ValidateDB(ctx, dbpool, "keyper", []string{
		"decryption_trigger",
		"decryption_key_share",
		"decryption_key",
	})
}
