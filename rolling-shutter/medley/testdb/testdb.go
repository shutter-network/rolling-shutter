package testdb

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprdb"
)

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

// NewTestDBPool connects to a test db specified an environment variable and clears it from all
// schemas we might have created. It returns the db connection pool and a close function. Call the
// close function at the end of the test to reset the db again and close the connection.
func NewTestDBPool(ctx context.Context, t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	testDBURL, exists := os.LookupEnv(testDBURLVar)
	if !exists {
		t.Skipf("no test db specified, please set %s", testDBURLVar)
	}

	dbpool, err := pgxpool.Connect(ctx, testDBURL)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	closedb := func() {
		_, err = dbpool.Exec(ctx, dropEverything)
		dbpool.Close() // close db no matter if dropping failed
		if err != nil {
			t.Fatalf("failed to reset test db: %v", err)
		}
	}

	// drop db contents
	_, err = dbpool.Exec(ctx, dropEverything)
	if err != nil {
		dbpool.Close()
		t.Fatalf("failed to reset test db: %v", err)
	}

	return dbpool, closedb
}

func NewKeyperTestDB(ctx context.Context, t *testing.T) (*kprdb.Queries, *pgxpool.Pool, func()) {
	t.Helper()

	dbpool, closedb := NewTestDBPool(ctx, t)
	db := kprdb.New(dbpool)
	err := kprdb.InitDB(ctx, dbpool)
	if err != nil {
		closedb()
		t.Fatalf("failed to initialize keyper db")
	}
	return db, dbpool, closedb
}

func NewCollatorTestDB(ctx context.Context, t *testing.T) (*cltrdb.Queries, *pgxpool.Pool, func()) {
	t.Helper()

	dbpool, closedb := NewTestDBPool(ctx, t)
	db := cltrdb.New(dbpool)
	err := cltrdb.InitDB(ctx, dbpool)
	if err != nil {
		log.Println(err)
		closedb()
		t.Fatalf("failed to initialize collator db")
	}
	return db, dbpool, closedb
}
