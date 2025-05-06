package syncmonitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

// MockSyncState is a mock implementation of BlockSyncState for testing.
type MockSyncState struct {
	mu          sync.Mutex
	blockNumber int64
	err         error
}

func (m *MockSyncState) GetSyncedBlockNumber(_ context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.blockNumber, m.err
}

func (m *MockSyncState) SetBlockNumber(n int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blockNumber = n
}

func TestSyncMonitor_ThrowsErrorWhenBlockNotIncreasing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialBlockNumber := int64(100)
	mockSyncState := &MockSyncState{
		blockNumber: initialBlockNumber,
	}

	monitor := &SyncMonitor{
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

	// Verify final state
	finalBlockNumber, err := mockSyncState.GetSyncedBlockNumber(ctx)
	assert.NilError(t, err)
	assert.Equal(t, initialBlockNumber, finalBlockNumber)
}

func TestSyncMonitor_HandlesBlockNumberIncreasing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialBlockNumber := int64(100)
	mockSyncState := &MockSyncState{
		blockNumber: initialBlockNumber,
	}

	monitor := &SyncMonitor{
		CheckInterval: 200 * time.Millisecond,
		SyncState:     mockSyncState,
	}

	monitorCtx, cancelMonitor := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	go func() {
		if err := service.RunWithSighandler(monitorCtx, monitor); err != nil {
			errCh <- err
		}
	}()

	// Update block numbers more quickly
	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
		mockSyncState.SetBlockNumber(initialBlockNumber + int64(i+1))
	}

	cancelMonitor()

	// Verify final state
	finalBlockNumber, err := mockSyncState.GetSyncedBlockNumber(ctx)
	assert.NilError(t, err)
	assert.Equal(t, initialBlockNumber+5, finalBlockNumber, "block number should have been incremented correctly")
}

func TestSyncMonitor_RunsNormallyWhenNoEons(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initialBlockNumber := int64(100)
	mockSyncState := &MockSyncState{
		blockNumber: initialBlockNumber,
	}

	monitor := &SyncMonitor{
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

	// Set up mock sync state that returns no rows error
	mockSyncState := &MockSyncState{
		err: pgx.ErrNoRows,
	}
	mockSyncState.SetBlockNumber(0) // Initialize block number

	monitor := &SyncMonitor{
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

func TestSyncMonitor_HandlesReorg(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up mock sync state that returns no rows error
	mockSyncState := &MockSyncState{}
	mockSyncState.SetBlockNumber(0) // Initialize block number

	monitor := &SyncMonitor{
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

	// Decrease the block number
	decreasedBlockNumber := int64(50)
	mockSyncState.SetBlockNumber(decreasedBlockNumber)

	time.Sleep(4 * time.Second)
	cancelMonitor()

	select {
	case err := <-errCh:
		t.Fatalf("expected monitor to continue without error, but got: %v", err)
	case <-time.After(1 * time.Second):
	}

	// Verify the block number was updated to the latest value
	syncedData, err := mockSyncState.GetSyncedBlockNumber(ctx)
	assert.NilError(t, err)
	assert.Equal(t, decreasedBlockNumber, syncedData, "block number should be updated to the decreased value")
}
