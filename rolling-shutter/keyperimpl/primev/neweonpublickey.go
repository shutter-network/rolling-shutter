package primev

import (
	"context"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
)

func (k *Keyper) processNewEonPublicKey(_ context.Context, key keyper.EonPublicKey) error { //nolint: unparam
	k.eonKeyPublisher.Publish(key)
	return nil
}
