package gnosis

import (
	"context"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
)

func (kpr *Keyper) processNewEonPublicKey(_ context.Context, key keyper.EonPublicKey) error { //nolint: unparam
	kpr.eonKeyPublisher.Publish(key)
	return nil
}
