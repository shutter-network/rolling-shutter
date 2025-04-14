package shutterservice_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"gotest.tools/assert"

	keyperDB "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func setupTestData(ctx context.Context, t *testing.T, dbpool *pgxpool.Pool, blockNumber int64) {
	t.Helper()
	db := database.New(dbpool)
	keyperdb := keyperDB.New(dbpool)

	// Set up eon
	err := keyperdb.InsertEon(ctx, keyperDB.InsertEonParams{
		Eon: 1,
	})
	assert.NilError(t, err)

	// Set up DKG result
	err = keyperdb.InsertDKGResult(ctx, keyperDB.InsertDKGResultParams{
		Eon:     1,
		Success: true,
	})
	assert.NilError(t, err)

	// Set up initial block
	err = db.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
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

func TestAPISyncMonitor_SkipsWhenDKGIsRunning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()
	db := database.New(dbpool)
	keyperdb := keyperDB.New(dbpool)

	// Set up eon but no DKG result to simulate DKG running
	err := keyperdb.InsertEon(ctx, keyperDB.InsertEonParams{
		Eon: 1,
	})
	assert.NilError(t, err)

	// Set up initial block data
	initialBlockNumber := int64(100)
	err = db.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
		BlockHash:   []byte{0x01, 0x02, 0x03},
		BlockNumber: initialBlockNumber,
	})
	assert.NilError(t, err)

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

	// Let it run for a while without incrementing the block number
	time.Sleep(15 * time.Second)
	cancelMonitor()

	select {
	case err := <-errCh:
		t.Fatalf("expected monitor to continue without error, but got: %v", err)
	case <-time.After(1 * time.Second):
		// Test passes if no error is received
	}

	// Verify the block number hasn't changed
	syncedData, err := db.GetIdentityRegisteredEventsSyncedUntil(ctx)
	assert.NilError(t, err)
	assert.Equal(t, initialBlockNumber, syncedData.BlockNumber, "block number should remain unchanged")
}

func TestAPISyncMonitor_RunsNormallyWhenNoEons(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()
	db := database.New(dbpool)

	// Only set up initial block data, no eon entries
	initialBlockNumber := int64(100)
	err := db.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
		BlockHash:   []byte{0x01, 0x02, 0x03},
		BlockNumber: initialBlockNumber,
	})
	assert.NilError(t, err)

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

	// Let it run for a while without incrementing the block number
	time.Sleep(15 * time.Second)
	cancelMonitor()

	select {
	case err := <-errCh:
		assert.ErrorContains(t, err, shutterservice.ErrBlockNotIncreasing.Error())
	case <-time.After(1 * time.Second):
		t.Fatalf("expected monitor to throw error, but no error returned")
	}

	// Verify the block number hasn't changed
	syncedData, err := db.GetIdentityRegisteredEventsSyncedUntil(ctx)
	assert.NilError(t, err)
	assert.Equal(t, initialBlockNumber, syncedData.BlockNumber, "block number should remain unchanged")
}

func TestAPISyncMonitor_ContinuesWhenNoRows(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()

	// Set up eon and DKG result, but no block data
	keyperdb := keyperDB.New(dbpool)

	err := keyperdb.InsertEon(ctx, keyperDB.InsertEonParams{
		Eon: 1,
	})
	assert.NilError(t, err)

	err = keyperdb.InsertDKGResult(ctx, keyperDB.InsertDKGResultParams{
		Eon:     1,
		Success: true,
	})
	assert.NilError(t, err)

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

func TestAPISyncMonitor_RunsNormallyWithCompletedDKG(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()
	db := database.New(dbpool)

	initialBlockNumber := int64(100)
	setupTestData(ctx, t, dbpool, initialBlockNumber)

	// Set up initial block data
	err := db.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
		BlockHash:   []byte{0x01, 0x02, 0x03},
		BlockNumber: initialBlockNumber,
	})
	assert.NilError(t, err)

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

	// Let it run for a while without incrementing the block number
	time.Sleep(15 * time.Second)
	cancelMonitor()

	select {
	case err := <-errCh:
		assert.ErrorContains(t, err, shutterservice.ErrBlockNotIncreasing.Error())
	case <-time.After(1 * time.Second):
		t.Fatalf("expected monitor to throw error, but no error returned")
	}

	// Verify the block number hasn't changed
	syncedData, err := db.GetIdentityRegisteredEventsSyncedUntil(ctx)
	assert.NilError(t, err)
	assert.Equal(t, initialBlockNumber, syncedData.BlockNumber, "block number should remain unchanged")
}
