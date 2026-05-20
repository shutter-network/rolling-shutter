package gnosis

import (
	"context"

	"github.com/rs/zerolog/log"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// processNewECIESKey upserts the keyper's ECIES public key into the local
// `ecies_keys` cache. It is invoked for both the synthetic events emitted by
// the initial poll and live `KeyRegistered` events from the registry contract.
func (kpr *Keyper) processNewECIESKey(ctx context.Context, ev *syncevent.ECIESKey) error {
	log.Debug().
		Str("keyper", ev.Keyper.Hex()).
		Int("key-len", len(ev.EciesPublicKey)).
		Msg("storing ECIES public key")
	return corekeyperdb.New(kpr.dbpool).UpsertECIESKey(ctx, corekeyperdb.UpsertECIESKeyParams{
		KeyperAddress:  shdb.EncodeAddress(ev.Keyper),
		EciesPublicKey: ev.EciesPublicKey,
	})
}
