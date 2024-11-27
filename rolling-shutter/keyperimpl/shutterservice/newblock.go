package shutterservice

import (
	"context"

	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
)

func (kpr *Keyper) processNewBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	if kpr.registrySyncer != nil {
		if err := kpr.registrySyncer.Sync(ctx, ev.Header); err != nil {
			return err
		}
	}
	return kpr.maybeTriggerDecryption(ctx)
}

// maybeTriggerDecryption triggers decryption for the identities registered if
// - it hasn't been triggered for thos identities before and
// - the keyper is part of the corresponding keyper set.
func (kpr *Keyper) maybeTriggerDecryption(ctx context.Context) error {
	//TODO: needs to be implemented
	return nil
}
