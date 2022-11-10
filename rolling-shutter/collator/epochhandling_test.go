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
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

func TestHandleDecryptionTriggerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, dbpool, closedb := testdb.NewCollatorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)

	// HACK: Only partially instantiating the collator.
	// This works until the handler/validator functions use something else than
	// the database-pool and the p2p handler
	// The reason the p2p.SendMessage works is because
	// the handler has no rooms subscribed and thus will actually
	// skip to forward the messages sent to the transport
	c := collator{dbpool: dbpool, Config: config, p2p: p2p.New(p2p.Config{})}

	trigger := cltrdb.InsertTriggerParams{
		EpochID:       epochid.Uint64ToEpochID(3).Bytes(),
		BatchHash:     common.BytesToHash([]byte{0, 1}).Bytes(),
		L1BlockNumber: 42,
	}
	err := db.InsertTrigger(ctx, trigger)
	assert.NilError(t, err)

	cctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err = c.handleNewDecryptionTrigger(cctx)
	assert.ErrorContains(t, err, "timeout: context deadline exceeded")

	triggers, err := db.GetUnsentTriggers(ctx)
	assert.NilError(t, err)

	assert.Equal(t, len(triggers), 0)
}
