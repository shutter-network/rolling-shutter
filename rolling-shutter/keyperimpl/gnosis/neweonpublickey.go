package gnosis

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
)

func (kpr *Keyper) processNewEonPublicKey(_ context.Context, key keyper.EonPublicKey) error { //nolint:unparam
	log.Info().
		Uint64("eon", key.Eon).
		Uint64("activation-block", key.ActivationBlock).
		Msg("new eon pk")
	return nil
}
