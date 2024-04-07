package gnosis

import (
	"context"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
)

func (kpr *Keyper) processNewEonPublicKey(ctx context.Context, key keyper.EonPublicKey) error {
	kpr.eonKeyPublisher.Publish(key)
	return nil
}
