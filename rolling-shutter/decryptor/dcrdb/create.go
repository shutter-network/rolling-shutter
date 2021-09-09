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

// InitDecryptorDB initializes the database of the decryptor. It is assumed that the db is empty.
func InitDecryptorDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	_, err := dbpool.Exec(ctx, CreateDecryptorTables)
	if err != nil {
		return errors.Wrap(err, "failed to create decryptor tables")
	}
	return nil
}

// ValidateDecryptorDB checks that all expected tables exist in the database. If not, it returns an
// error.
func ValidateDecryptorDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return shdb.ValidateDB(ctx, dbpool, "decryptor", []string{
		"cipher_batch",
		"decryption_key",
		"decryption_signature",
	})
}
