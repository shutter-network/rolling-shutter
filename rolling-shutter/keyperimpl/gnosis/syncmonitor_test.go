package gnosis_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func TestSyncMonitor_ThrowsErrorWhenBlockNotIncreasing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()

	_, err := dbpool.Exec(ctx, `
    CREATE TABLE IF NOT EXISTS transaction_submitted_events_synced_until(
        enforce_one_row bool PRIMARY KEY DEFAULT true,
        block_hash bytea NOT NULL,
        block_number bigint NOT NULL CHECK (block_number >= 0),
        slot bigint NOT NULL CHECK (slot >= 0)
    );
    `)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	initialBlockNumber := int64(100)
	_, err = dbpool.Exec(ctx, `
    INSERT INTO transaction_submitted_events_synced_until (block_hash, block_number, slot)
    VALUES ($1, $2, $3);
    `, []byte{0x01, 0x02, 0x03}, initialBlockNumber, 1)
	if err != nil {
		t.Fatalf("failed to insert initial data: %v", err)
	}

	monitor := &gnosis.SyncMonitor{
		DBPool: dbpool,
	}

	errCh := make(chan error, 1)

	go func() {
		err := service.RunWithSighandler(ctx, monitor)
		if err != nil {
			errCh <- err
		}
	}()

	time.Sleep(80 * time.Second)

	select {
	case err := <-errCh:
		assert.ErrorContains(t, err, "block number has not increased between checks")
	case <-time.After(5 * time.Second):
		t.Fatal("expected an error, but none was returned")
	}
}

func TestSyncMonitor_HandlesBlockNumberIncreasing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()

	_, err := dbpool.Exec(ctx, `
    CREATE TABLE IF NOT EXISTS transaction_submitted_events_synced_until(
        enforce_one_row bool PRIMARY KEY DEFAULT true,
        block_hash bytea NOT NULL,
        block_number bigint NOT NULL CHECK (block_number >= 0),
        slot bigint NOT NULL CHECK (slot >= 0)
    );
`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	initialBlockNumber := int64(100)
	_, err = dbpool.Exec(ctx, `
    INSERT INTO transaction_submitted_events_synced_until (block_hash, block_number, slot)
    VALUES ($1, $2, $3);
`, []byte{0x01, 0x02, 0x03}, initialBlockNumber, 1)
	if err != nil {
		t.Fatalf("failed to insert initial data: %v", err)
	}

	var count int
	err = dbpool.QueryRow(ctx, `
    SELECT count(*) FROM transaction_submitted_events_synced_until
    WHERE block_number = $1;
`, initialBlockNumber).Scan(&count)
	if err != nil {
		t.Fatalf("failed to verify initial data: %v", err)
	}

	assert.Equal(t, 1, count, "initial data should be inserted")

	monitor := &gnosis.SyncMonitor{
		DBPool: dbpool,
	}

	_, deferFn := service.RunBackground(ctx, monitor)
	defer deferFn()

	doneCh := make(chan struct{})
	go func() {
		for i := 0; i < 5; i++ {
			// Simulate block number increment by updating the database
			newBlockNumber := initialBlockNumber + int64(i+1)
			log.Info().
				Int64("previous-block-number", initialBlockNumber+int64(i)).
				Int64("new-block-number", newBlockNumber).
				Msg("comparing consecutive blocks")

			_, err := dbpool.Exec(ctx, `
        UPDATE transaction_submitted_events_synced_until
        SET block_number = $1
        WHERE block_number = $2;
`, newBlockNumber, initialBlockNumber+int64(i))
			if err != nil {
				t.Errorf("failed to update block number: %v", err)
				return
			}

			time.Sleep(30 * time.Second)
		}

		doneCh <- struct{}{}
	}()

	<-doneCh
	var finalBlockNumber int64
	err = dbpool.QueryRow(ctx, `SELECT block_number FROM transaction_submitted_events_synced_until LIMIT 1;`).Scan(&finalBlockNumber)
	if err != nil {
		t.Fatalf("failed to retrieve final block number: %v", err)
	}

	assert.Equal(t, initialBlockNumber+5, finalBlockNumber, "block number should have been incremented correctly")
}
