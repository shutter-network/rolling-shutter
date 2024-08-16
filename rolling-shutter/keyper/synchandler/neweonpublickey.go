package gnosis

import (
	"context"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
)

// TODO: put in the eventhandler directly
func processNewEonPublicKey(_ context.Context, key keyper.EonPublicKey) error { //nolint: unparam
	kpr.eonKeyPublisher.Publish(key)
	return nil
}
