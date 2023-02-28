package db

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

//go:embed */schema.sql
var schemas embed.FS

func GetSchema(n string) string {
	b, err := schemas.ReadFile(n + "/schema.sql")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func MustFindSchemaVersion(path string) string {
	return shdb.MustFindSchemaVersion(GetSchema(path), path)
}

func Create(ctx context.Context, tx pgx.Tx, paths []string) error {
	for _, p := range paths {
		s := GetSchema(p)
		_, err := tx.Exec(ctx, s)
		if err != nil {
			return errors.Wrapf(err, "failed to execute SQL statements in %s", p)
		}
	}
	return nil
}
