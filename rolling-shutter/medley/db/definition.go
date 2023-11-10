package db

import (
	"embed"

	"github.com/rs/zerolog/log"
)

//go:generate sqlc generate --file sql/sqlc.yaml

//go:embed sql
var files embed.FS
var metaDefinition Definition

func init() {
	var err error
	metaDefinition, err = NewSQLCDefinition(files, "sql/", "meta", 1)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize DB metadata")
	}
}
