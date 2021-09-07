package shdb

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

// tableNamesQuery returns the names of all user created tables in the database.
const tableNamesQuery = `
	SELECT table_name
	FROM information_schema.tables
	WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
	AND table_schema NOT LIKE 'pg_toast%'
`

// dbNotEmptyQuery returns one row if there is at least one user created table in the database,
// otherwise no row.
const dbNotEmptyQuery = `
	SELECT 1
	FROM information_schema.tables
	WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
	AND table_schema NOT LIKE 'pg_toast%'
	LIMIT 1;
`

// ValidateKeyperDB checks that all expected tables exist in the database. If not, it returns an
// error.
func ValidateKeyperDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return validateDB(ctx, dbpool, []string{
		"decryption_trigger",
		"decryption_key_share",
		"decryption_key",
	})
}

// ValidateDecryptorDB checks that all expected tables exist in the database. If not, it returns an
// error.
func ValidateDecryptorDB(ctx context.Context, dbpool *pgxpool.Pool) error {
	return validateDB(ctx, dbpool, []string{
		"cipher_batch",
		"decryption_key",
		"decryption_signature",
	})
}

func validateDB(ctx context.Context, dbpool *pgxpool.Pool, requiredTables []string) error {
	requiredTableMap := make(map[string]bool)
	for _, table := range requiredTables {
		requiredTableMap[table] = true
	}

	rows, err := dbpool.Query(ctx, tableNamesQuery)
	if err != nil {
		return errors.Wrap(err, "failed to query table names from db")
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		err := rows.Scan(&tableName)
		if err != nil {
			return errors.Wrap(err, "failed to query table names from db")
		}
		delete(requiredTableMap, tableName)
	}
	if rows.Err() != nil {
		return errors.Wrap(rows.Err(), "read table names")
	}

	if len(requiredTableMap) != 0 {
		return errors.New("database misses one or more required table")
	}
	return nil
}

// IsEmpty checks if the db contains no non-default tables.
func IsEmpty(ctx context.Context, dbpool *pgxpool.Pool) (bool, error) {
	rows, err := dbpool.Query(ctx, dbNotEmptyQuery)
	if err != nil {
		return false, errors.Wrap(err, "failed to ask db if it is empty")
	}
	defer rows.Close()

	if rows.Next() {
		return false, nil // got row, no error
	}
	err = rows.Err()
	if err != nil {
		return true, errors.Wrap(err, "failed to get row from query") // no row, but error
	}
	return true, nil // no row, no error
}
