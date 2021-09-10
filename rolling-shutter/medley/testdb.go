package medley

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
)

const testDBURLVar = "ROLLING_SHUTTER_TESTDB_URL"

const dropEverything = `
	DROP SCHEMA IF EXISTS decryptor CASCADE;
	DROP SCHEMA IF EXISTS keyper CASCADE;
`

// NewTestDBPool connects to a test db specified an environment variable and clears it from all
// schemas we might have created. It returns the db connection pool and a close function. Call the
// close function at the end of the test to reset the db again and close the connection.
func NewTestDBPool(ctx context.Context, t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	testDBURL := os.Getenv(testDBURLVar)
	if testDBURL == "" {
		t.Fatal("no test db specified")
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

func NewKeyperTestDB(ctx context.Context, t *testing.T) (*kprdb.Queries, func()) {
	t.Helper()

	dbpool, closedb := NewTestDBPool(ctx, t)
	db := kprdb.New(dbpool)
	err := kprdb.InitKeyperDB(ctx, dbpool)
	if err != nil {
		closedb()
		t.Fatalf("failed to initialize keyper db")
	}
	return db, closedb
}

func NewDecryptorTestDB(ctx context.Context, t *testing.T) (*dcrdb.Queries, func()) {
	t.Helper()

	dbpool, closedb := NewTestDBPool(ctx, t)
	db := dcrdb.New(dbpool)
	err := dcrdb.InitDecryptorDB(ctx, dbpool)
	if err != nil {
		closedb()
		t.Fatalf("failed to initialize decryptor db")
	}
	return db, closedb
}
