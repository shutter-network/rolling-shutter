package commondb

import (
	_ "embed"
)

// CreateObserveTables contains the SQL statements to create observe tables.
//
//go:embed observe-schema.sql
var CreateObserveTables string
