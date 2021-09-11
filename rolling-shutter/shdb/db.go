package shdb

import (
	"log"
	"regexp"
)

const SchemaVersionKey = "schema-version"

// MustFindSchemaVersion extracts the expected schema version from schema.sql files that should
// contain a header like:
// -- schema-version: 23 --
// It aborts the program, if the version cannot be determined.
func MustFindSchemaVersion(schema, path string) string {
	rx := "-- schema-version: ([0-9]+) --"
	matches := regexp.MustCompile(rx).FindStringSubmatch(schema)
	if len(matches) != 2 {
		log.Fatalf("malformed schema in %s, cannot find regular expression %s", path, rx)
	}
	return matches[1]
}
