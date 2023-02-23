package chainobsdb

import (
	_ "embed"
)

// CreateObserveTables contains the SQL statements to create observe tables.
//
//go:embed schema.sql
var CreateObserveTables string
