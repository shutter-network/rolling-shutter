package collator

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testlog"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func init() {
	testlog.Setup()
}

func TestHandleDecryptionTriggerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, dbpool, closedb := testdb.NewCollatorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)

	// HACK: Only partially instantiating the collator.
	// This works until the handler/validator functions use something else than
	// the database-pool and the p2p handler
	// The reason the p2p.SendMessage works without configuration
	// is because the handler has no rooms subscribed and thus will actually
	// skip to forward the messages sent to the transport
	c := collator{dbpool: dbpool, Config: config, p2p: p2p.New(p2p.Config{})}
	var newDecryptionTriggerLoop shdb.SignalLoopFunc
	c.signals.newDecryptionTrigger, newDecryptionTriggerLoop = shdb.NewSignal(ctx, "newDecryptionTrigger", c.sendDecryptionTriggers)
	go func() { _ = newDecryptionTriggerLoop() }()

	trigger := cltrdb.InsertTriggerParams{
		EpochID:       epochid.Uint64ToEpochID(3).Bytes(),
		BatchHash:     common.BytesToHash([]byte{0, 1}).Bytes(),
		L1BlockNumber: 42,
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		err := db.InsertTrigger(ctx, trigger)
		assert.NilError(t, err)
	}()

	cctx, cancelTimeout := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancelTimeout()
	err := c.handleDatabaseNotifications(cctx)
	assert.ErrorContains(t, err, "context deadline exceeded")

	// HACK: The handleNewDecryptionTrigger should have
	// handed the messages to the P2P-messaging (see above) and marked it as sent
	// in the DB. Thus the following query doesn't return any triggers:
	triggers, err := db.GetUnsentTriggers(ctx)
	assert.NilError(t, err)

	assert.Equal(t, len(triggers), 0)
}
