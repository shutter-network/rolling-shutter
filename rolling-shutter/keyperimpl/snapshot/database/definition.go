package database

import (
	chainobsdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
)

var Definition = db.NewAggregateDefinition(
	"snapshot",
	database.Definition,
	chainobsdb.CollatorDefinition,
)
