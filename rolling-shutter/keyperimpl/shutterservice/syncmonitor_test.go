package shutterservice

import (
	"context"
	"testing"
	"time"

	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func TestSyncMonitor_ThrowsErrorWhenBlockNotIncreasing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, dbclose := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer dbclose()
	db := database.New(dbpool)

	initialBlockNumber := int64(100)

	err := db.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
		BlockHash:   []byte{0x01, 0x02, 0x03},
		BlockNumber: initialBlockNumber,
	})
	if err != nil {
		t.Fatalf("failed to set initial synced data: %v", err)
	}

	monitor := &SyncMonitor{
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
	db := database.New(dbpool)

	initialBlockNumber := int64(100)
	err := db.SetIdentityRegisteredEventSyncedUntil(ctx, database.SetIdentityRegisteredEventSyncedUntilParams{
		BlockHash:   []byte{0x01, 0x02, 0x03},
		BlockNumber: initialBlockNumber,
	})
	if err != nil {
		t.Fatalf("failed to set initial synced data: %v", err)
	}

	monitor := &SyncMonitor{
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

func TestSyncMonitor_ContinuesWhenNoRows(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool, closeDB := testsetup.NewTestDBPool(ctx, t, database.Definition)
	defer closeDB()
	_ = database.New(dbpool)

	monitor := &SyncMonitor{
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
