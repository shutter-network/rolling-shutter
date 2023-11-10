package database

import (
	collatordb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/collator"
	keyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
)

var (
	KeyperDefinition   db.Definition = keyperdb.Definition
	CollatorDefinition db.Definition = collatordb.Definition
)
