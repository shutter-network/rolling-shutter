package gnosis

import (
	"context"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
)

func (kpr *Keyper) processNewBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	// The sequencer + validator syncers used to fetch events via
	// eth_getLogs on every new block. That per-block poll has been replaced
	// by WatchTransactionSubmitted / WatchUpdated subscriptions set up in
	// their Start methods (C4). processNewBlock now only drives the
	// slot-timed decryption trigger.
	slot := medley.BlockTimestampToSlot(
		ev.Header.Time,
		kpr.config.Gnosis.GenesisSlotTimestamp,
		kpr.config.Gnosis.SecondsPerSlot,
	)
	return kpr.maybeTriggerDecryption(ctx, slot+1)
}
