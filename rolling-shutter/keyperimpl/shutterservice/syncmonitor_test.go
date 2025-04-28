package shutterservice_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func setupTestData(ctx context.Context, t *testing.T, dbpool *pgxpool.Pool, blockNumber int64) {
	t.Helper()
	db := database.New(dbpool)

	// Set up initial block
	err := db.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
		BlockHash:   []byte{0x01, 0x02, 0x03},
		BlockNumber: blockNumber,
	})
	assert.NilError(t, err)
}

func TestAPISyncMonitor_ThrowsErrorWhenBlockNotIncreasing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer dbclose()

	initialBlockNumber := int64(100)
	setupTestData(ctx, t, dbpool, initialBlockNumber)

	monitor := &shutterservice.SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		err := service.RunWithSighandler(ctx, monitor)
		if err != nil {
			errCh <- err
		}
	}()

	time.Sleep(12 * time.Second)

	select {
	case err := <-errCh:
		assert.ErrorContains(t, err, shutterservice.ErrBlockNotIncreasing.Error())
	case <-time.After(5 * time.Second):
		t.Fatal("expected an error, but none was returned")
	}
}

func TestAPISyncMonitor_HandlesBlockNumberIncreasing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()
	db := database.New(dbpool)

	initialBlockNumber := int64(100)
	setupTestData(ctx, t, dbpool, initialBlockNumber)

	monitor := &shutterservice.SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
	}

	_, deferFn := service.RunBackground(ctx, monitor)
	defer deferFn()

	doneCh := make(chan struct{})
	go func() {
		for i := 0; i < 5; i++ {
			newBlockNumber := initialBlockNumber + int64(i+1)
			err := db.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
				BlockHash:   []byte{0x01, 0x02, 0x03},
				BlockNumber: newBlockNumber,
			})
			if err != nil {
				t.Errorf("failed to update block number: %v", err)
				return
			}

			time.Sleep(5 * time.Second)
		}

		doneCh <- struct{}{}
	}()

	<-doneCh
	syncedData, err := db.GetIdentityRegisteredEventsSyncedUntil(ctx)
	if err != nil {
		t.Fatalf("failed to retrieve final block number: %v", err)
	}

	assert.Equal(t, initialBlockNumber+5, syncedData.BlockNumber, "block number should have been incremented correctly")
}

func TestAPISyncMonitor_ContinuesWhenNoRows(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()

	monitor := &shutterservice.SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
	}

	monitorCtx, cancelMonitor := context.WithCancel(ctx)
	defer cancelMonitor()

	errCh := make(chan error, 1)
	go func() {
		err := service.RunWithSighandler(monitorCtx, monitor)
		if err != nil {
			errCh <- err
		}
	}()

	time.Sleep(15 * time.Second)
	cancelMonitor()

	select {
	case err := <-errCh:
		t.Fatalf("expected monitor to continue without error, but got: %v", err)
	case <-time.After(1 * time.Second):
	}
}

func TestAPISyncMonitor_HandlesReorg(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()
	db := database.New(dbpool)

	// Set up initial block at a higher number
	initialBlockNumber := int64(100)
	setupTestData(ctx, t, dbpool, initialBlockNumber)

	monitor := &shutterservice.SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
	}

	monitorCtx, cancelMonitor := context.WithCancel(ctx)
	defer cancelMonitor()

	errCh := make(chan error, 1)
	go func() {
		err := service.RunWithSighandler(monitorCtx, monitor)
		if err != nil {
			errCh <- err
		}
	}()

	// Decrease the block number
	decreasedBlockNumber := int64(50)
	err := db.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
		BlockHash:   []byte{0x01, 0x02, 0x03},
		BlockNumber: decreasedBlockNumber,
	})
	assert.NilError(t, err)

	time.Sleep(4 * time.Second)
	cancelMonitor()

	select {
	case err := <-errCh:
		t.Fatalf("expected monitor to continue without error, but got: %v", err)
	case <-time.After(1 * time.Second):
	}

	// Verify the block number was updated to the latest value
	syncedData, err := db.GetIdentityRegisteredEventsSyncedUntil(ctx)
	assert.NilError(t, err)
	assert.Equal(t, decreasedBlockNumber, syncedData.BlockNumber, "block number should be updated to the decreased value")
}
