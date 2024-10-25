package gnosis

import (
	"context"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/legacychainsync/event"
)

func (kpr *Keyper) processNewBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	if kpr.sequencerSyncer != nil {
		if err := kpr.sequencerSyncer.Sync(ctx, ev.Header); err != nil {
			return err
		}
	}
	err := kpr.validatorSyncer.Sync(ctx, ev.Header)
	if err != nil {
		return err
	}
	slot := medley.BlockTimestampToSlot(
		ev.Header.Time,
		kpr.config.Gnosis.GenesisSlotTimestamp,
		kpr.config.Gnosis.SecondsPerSlot,
	)
	return kpr.maybeTriggerDecryption(ctx, slot+1)
}
