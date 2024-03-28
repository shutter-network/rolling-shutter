package gnosis

import (
	"context"

	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
)

func (kpr *Keyper) processNewBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	if kpr.sequencerSyncer != nil {
		if err := kpr.sequencerSyncer.Sync(ctx, ev.Header); err != nil {
			return err
		}
	}
	return nil
}
