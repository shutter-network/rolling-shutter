package syncmonitor

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

// MockSyncState is a mock implementation of BlockSyncState for testing
type MockSyncState struct {
	blockNumber int64
	err         error
}

func (m *MockSyncState) GetSyncedBlockNumber(ctx context.Context) (int64, error) {
	return m.blockNumber, m.err
}

func setupTestData(ctx context.Context, t *testing.T, dbpool *pgxpool.Pool) {
	t.Helper()
	keyperdb := database.New(dbpool)

	// Set up eon
	err := keyperdb.InsertEon(ctx, database.InsertEonParams{
		Eon: 1,
	})
	assert.NilError(t, err)

	// Set up DKG result
	err = keyperdb.InsertDKGResult(ctx, database.InsertDKGResultParams{
		Eon:     1,
		Success: true,
	})
	assert.NilError(t, err)
}

func TestSyncMonitor_ThrowsErrorWhenBlockNotIncreasing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer dbclose()

	initialBlockNumber := int64(100)
	setupTestData(ctx, t, dbpool)

	mockSyncState := &MockSyncState{
		blockNumber: initialBlockNumber,
	}

	monitor := &SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
		SyncState:     mockSyncState,
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
		assert.ErrorContains(t, err, ErrBlockNotIncreasing.Error())
	case <-time.After(5 * time.Second):
		t.Fatal("expected an error, but none was returned")
	}
}

func TestSyncMonitor_HandlesBlockNumberIncreasing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()

	initialBlockNumber := int64(100)
	setupTestData(ctx, t, dbpool)

	mockSyncState := &MockSyncState{
		blockNumber: initialBlockNumber,
	}

	monitor := &SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
		SyncState:     mockSyncState,
	}

	_, deferFn := service.RunBackground(ctx, monitor)
	defer deferFn()

	doneCh := make(chan struct{})
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(5 * time.Second)
			mockSyncState.blockNumber = initialBlockNumber + int64(i+1)
		}

		doneCh <- struct{}{}
	}()

	<-doneCh
	assert.Equal(t, initialBlockNumber+5, mockSyncState.blockNumber, "block number should have been incremented correctly")
}

func TestSyncMonitor_SkipsWhenDKGIsRunning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()
	keyperdb := database.New(dbpool)

	// Set up eon but no DKG result to simulate DKG running
	err := keyperdb.InsertEon(ctx, database.InsertEonParams{
		Eon: 1,
	})
	assert.NilError(t, err)

	initialBlockNumber := int64(100)
	mockSyncState := &MockSyncState{
		blockNumber: initialBlockNumber,
	}

	monitor := &SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
		SyncState:     mockSyncState,
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
	assert.Equal(t, initialBlockNumber, mockSyncState.blockNumber, "block number should remain unchanged")
}

func TestSyncMonitor_RunsNormallyWhenNoEons(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()

	initialBlockNumber := int64(100)
	mockSyncState := &MockSyncState{
		blockNumber: initialBlockNumber,
	}

	monitor := &SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
		SyncState:     mockSyncState,
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
		assert.ErrorContains(t, err, ErrBlockNotIncreasing.Error())
	case <-time.After(1 * time.Second):
		t.Fatalf("expected monitor to throw error, but no error returned")
	}
}

func TestSyncMonitor_ContinuesWhenNoRows(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()

	// Set up eon and DKG result
	keyperdb := database.New(dbpool)
	err := keyperdb.InsertEon(ctx, database.InsertEonParams{
		Eon: 1,
	})
	assert.NilError(t, err)

	err = keyperdb.InsertDKGResult(ctx, database.InsertDKGResultParams{
		Eon:     1,
		Success: true,
	})
	assert.NilError(t, err)

	// Set up mock sync state that returns no rows error
	mockSyncState := &MockSyncState{
		err: pgx.ErrNoRows,
	}

	monitor := &SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
		SyncState:     mockSyncState,
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
		// Test passes if no error is received
	}
}

func TestSyncMonitor_RunsNormallyWithCompletedDKG(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()

	setupTestData(ctx, t, dbpool)

	initialBlockNumber := int64(100)
	mockSyncState := &MockSyncState{
		blockNumber: initialBlockNumber,
	}

	monitor := &SyncMonitor{
		DBPool:        dbpool,
		CheckInterval: 5 * time.Second,
		SyncState:     mockSyncState,
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
		assert.ErrorContains(t, err, ErrBlockNotIncreasing.Error())
	case <-time.After(1 * time.Second):
		t.Fatalf("expected monitor to throw error, but no error returned")
	}
}
