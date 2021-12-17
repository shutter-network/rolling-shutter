package commondb

import (
	_ "embed"
)

//go:embed meta-schema.sql
// CreateMetaInf contains the SQL statements to create the meta_inf table.
var CreateMetaInf string

//go:embed observe-schema.sql
// CreateObserveTables contains the SQL statements to create observe tables.
var CreateObserveTables string
