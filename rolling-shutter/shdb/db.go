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
	WHERE table_schema = $1
`

func ValidateDB(ctx context.Context, dbpool *pgxpool.Pool, schema string, requiredTables []string) error {
	requiredTableMap := make(map[string]bool)
	for _, table := range requiredTables {
		requiredTableMap[table] = true
	}

	rows, err := dbpool.Query(ctx, tableNamesQuery, schema)
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
