package primev

import (
	"context"

	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
)

func (k *Keyper) processNewBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	// TODO: need to fetch new provider registrations
	return nil
}
