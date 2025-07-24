package database

import (
	"embed"

	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
)

//go:generate sqlc generate --file sql/sqlc.yaml

// TODO: add the sql files here
var files embed.FS

var Definition db.Definition

func init() {
	def, err := db.NewSQLCDefinition(files, "sql/", "primevkeyper", 1)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize DB metadata")
	}
	Definition = db.NewAggregateDefinition(
		"primevkeyper",
		def,
		database.Definition,
	)
}
