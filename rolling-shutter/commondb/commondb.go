package commondb

import (
	_ "embed"
)

//go:embed meta-schema.sql
// CreateMetaInf contains the SQL statements to create the meta_info table.
var CreateMetaInf string
