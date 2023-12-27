package epochkghandler

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type KeyShareHandler struct {
	InstanceID    uint64
	KeyperAddress common.Address
	DBPool        *pgxpool.Pool

	Messaging p2p.Messaging
	Trigger   <-chan *broker.Event[*DecryptionTrigger]
}

func (ksh *KeyShareHandler) listen(ctx context.Context) error {
	for {
		select {
		case triggerEvent, ok := <-ksh.Trigger:
			if !ok {
				log.Debug().Msg("decryption trigger channel closed, stopping loop")
				return nil
			}
			metricsEpochKGDecryptionTriggersReceived.Inc()
			trigger := triggerEvent.Value
			eon, err := ksh.getEonForBlockNumber(ctx, trigger.BlockNumber)
			if err != nil {
				log.Error().Err(err).Msg("error retrieving eon for blocknumber")
				// FIXME:how to handle
				continue
			}
			log.Debug().Msg("constructing dectryption key share")
			keySharesMsg, err := ksh.ConstructDecryptionKeyShare(
				ctx,
				eon,
				trigger.IdentityPreimages,
			)
			if keySharesMsg == nil {
				// FIXME: how to handle
				continue
			}
			if err != nil {
				// FIXME: how to handle
				continue
			}

			log.Debug().Msg("sending key share message")
			// TODO: retry opts
			if err := ksh.Messaging.SendMessage(ctx, keySharesMsg); err != nil {
				log.Error().Err(err).Msg("error while sending p2p message")
				// FIXME: how to handle?
				continue
			}
			metricsEpochKGDecryptionKeySharesSent.Inc()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (ksh *KeyShareHandler) Start(ctx context.Context, group service.Runner) error {
	group.Go(func() error {
		return ksh.listen(ctx)
	})
	return nil
}
