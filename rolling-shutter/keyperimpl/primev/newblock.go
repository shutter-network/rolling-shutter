package primev

import (
	"context"

	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
)

// TODO: the syncing logic for provider registry to be stripped out of here.
//
//	As it uses different chain. Adding it here, just for convenience of development of POC.
func (k *Keyper) processNewBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	if k.providerRegistrySyncer != nil {
		if err := k.providerRegistrySyncer.Sync(ctx, ev.Header); err != nil {
			return err
		}
	}
	return nil
}
