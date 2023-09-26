package db

//go:generate sqlc generate

import (
	"context"
	"embed"
	stdPath "path"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

var pathToSchema map[string][]string

func init() {
	var err error
	pathToSchema, err = getMap()
	if err != nil {
		// TODO error message
		panic(err)
	}
	// NOCHECKIN
	log.Info().Interface("Schemas", pathToSchema).Msg("")
}

type Data struct {
	SQL []Entry `yaml:"sql"`
}

type Entry struct {
	Path    string   `yaml:"gen.go.out,flow"`
	Schemas []string `yaml:"schema"`
}

//go:embed sqlc.yaml */sql/*/schema.sql
var sqlSchemas embed.FS

func GetSQLCreateStatements(path string) []string {
	schemaPaths := pathToSchema[path]
	sqlStatements := []string{}
	for _, n := range schemaPaths {
		b, err := sqlSchemas.ReadFile(n + "/schema.sql")
		if err != nil {
			panic(err)
		}

		sqlStatements = append(sqlStatements, string(b))
	}
	return sqlStatements
}

func MustFindSchemaVersion(path string) string {
	schemaToVersion := map[string]string{}
	for _, sqlPath := range GetSQLCreateStatements(path) {
		versionString := shdb.MustFindSchemaVersion(sqlPath, path)
		schema, _ := strings.CutSuffix(stdPath.Base(path), ".sql")
		schemaToVersion[schema] = versionString
	}
	// TODO Now serialize deterministically
	// FIXME find a version of this that works with
	// multiple sql files and thus multiple schema versions
	// the below is not comleted yet
	return ""
}

func getMap() (map[string][]string, error) {
	m := make(map[string][]string)
	b, err := sqlSchemas.ReadFile("sqlc.yaml")
	if err != nil {
		return m, err
	}
	var data Data
	err = yaml.Unmarshal(b, &data)
	if err != nil {
		return m, err
	}
	for _, entry := range data.SQL {
		m[entry.Path] = entry.Schemas
	}
	log.Info().Interface("data", data).Msg("parsed sqlc.yaml")
	return m, nil
}

func Create(ctx context.Context, tx pgx.Tx, paths []string) error {
	for _, p := range paths {
		sqlStatements := GetSQLCreateStatements(p)
		for _, s := range sqlStatements {
			_, err := tx.Exec(ctx, s)
			if err != nil {
				return errors.Wrapf(err, "failed to execute SQL statements in %s", p)
			}
		}
	}
	return nil
}
