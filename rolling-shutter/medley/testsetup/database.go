package testsetup

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testlog"
)

func init() {
	testlog.Setup()
}

const testDBURLVar = "ROLLING_SHUTTER_TESTDB_URL"

const dropEverything = `
DO $$ DECLARE
    r RECORD;
BEGIN
    -- if the schema you operate on is not "current", you will want to
    -- replace current_schema() in query with 'schematodeletetablesfrom'
    -- *and* update the generate 'DROP...' accordingly.
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
	EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;

    FOR r in (select typname from pg_type where typelem=0 AND typnamespace IN (select oid from pg_namespace where nspname=current_schema()))
    LOOP
      EXECUTE 'DROP TYPE IF EXISTS ' || quote_ident(r.typname);
    END LOOP;

END $$;
`

var testDBSuffix = "-test"

// newDBPoolTeardown connects to a test db specified an environment variable and clears it from all
// schemas we might have created. It returns the db connection pool and a close function. Call the
// close function at the end of the test to reset the db again and close the connection.
func newDBPoolTeardown(ctx context.Context, tb testing.TB) (*pgxpool.Pool, func()) {
	tb.Helper()

	testDBURL, exists := os.LookupEnv(testDBURLVar)
	if !exists {
		tb.Skipf("no test db specified, please set %s", testDBURLVar)
	}

	dbpool, err := pgxpool.Connect(ctx, testDBURL)
	if err != nil {
		tb.Fatalf("failed to connect to test db: %v", err)
	}

	closedb := func() {
		_, err = dbpool.Exec(ctx, dropEverything)
		dbpool.Close() // close db no matter if dropping failed
		if err != nil {
			tb.Fatalf("failed to reset test db: %v", err)
		}
	}

	// drop db contents
	_, err = dbpool.Exec(ctx, dropEverything)
	if err != nil {
		dbpool.Close()
		tb.Fatalf("failed to reset test db: %v", err)
	}

	return dbpool, closedb
}

func NewTestDBPool(ctx context.Context, tb testing.TB, definition db.Definition) (*pgxpool.Pool, func()) {
	tb.Helper()

	dbpool, closedb := newDBPoolTeardown(ctx, tb)

	err := db.InitDB(ctx, dbpool, definition.Name()+testDBSuffix, definition)
	if err != nil {
		log.Error().Err(err).Str("db-definition", definition.Name()).Msg("Initializing DB failed")
		closedb()
		tb.Fatalf("failed to initialize '%s' db", definition.Name())
	}
	return dbpool, closedb
}
