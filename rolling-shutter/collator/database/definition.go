package database

import (
	"embed"

	"github.com/rs/zerolog/log"

	chainobsdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
)

//go:generate sqlc generate --file sql/sqlc.yaml

//go:embed sql
var Files embed.FS
var Definition db.Definition

func init() {
	def, err := db.NewSQLCDefinition(Files, "sql/", "collator", 1)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize DB")
	}
	Definition = db.NewAggregateDefinition(
		"collator",
		def,
		chainobsdb.KeyperDefinition,
		chainobsdb.CollatorDefinition,
	)
}
