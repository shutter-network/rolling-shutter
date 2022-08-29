package commondb

import (
	_ "embed"
)

// CreateMetaInf contains the SQL statements to create the meta_inf table.
//
//go:embed meta-schema.sql
var CreateMetaInf string

// CreateObserveTables contains the SQL statements to create observe tables.
//
//go:embed observe-schema.sql
var CreateObserveTables string
