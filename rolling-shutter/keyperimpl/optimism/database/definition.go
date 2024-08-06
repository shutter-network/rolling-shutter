package database

import (
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
)

var Definition = db.NewAggregateDefinition(
	"opkeyper",
	database.Definition,
)
