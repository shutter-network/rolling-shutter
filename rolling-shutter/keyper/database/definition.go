package database

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"

	chainobsdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
)

//go:generate sqlc generate --file sql/sqlc.yaml

//go:embed sql
var files embed.FS

var Definition db.Definition

func init() {
	sqlcDB, err := db.NewSQLCDefinition(files, "sql/", "keyper")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize DB")
	}
	Definition = db.NewAggregateDefinition(
		"keyper",
		&KeyperDB{sqlcDB},
		chainobsdb.KeyperDefinition,
	)
}

// Wraps the SQLCDatabase for custom init logic.
type KeyperDB struct {
	db.Definition
}

func (d *KeyperDB) Name() string {
	return d.Definition.Name()
}

func (d *KeyperDB) Create(ctx context.Context, tx pgx.Tx) error {
	return d.Definition.Create(ctx, tx)
}

func (d *KeyperDB) Init(ctx context.Context, tx pgx.Tx) error {
	return d.Definition.Init(ctx, tx)
}

func (d *KeyperDB) Validate(ctx context.Context, tx pgx.Tx) error {
	return d.Definition.Validate(ctx, tx)
}
