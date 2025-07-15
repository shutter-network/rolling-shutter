package epochkghandler

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/retry"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

var DefaultKeyShareHandlerRetryOpts = []retry.Option{}

type KeyShareHandler struct {
	InstanceID           uint64
	KeyperAddress        common.Address
	MaxNumKeysPerMessage uint64
	DBPool               *pgxpool.Pool

	Messaging p2p.Messaging
	Trigger   <-chan *broker.Event[*DecryptionTrigger]
}

func (ksh *KeyShareHandler) handleEvent(ctx context.Context, ev *broker.Event[*DecryptionTrigger]) {
	var err error
	defer func() {
		if err != nil {
			if errors.Is(err, ErrIgnoreDecryptionRequest) {
				log.Debug().Err(errors.Unwrap(err)).Msg("ignoring request for key share release")
				//  set the event result to OK=true here, e.g. so that the sender doesn't reschedule
				_ = ev.SetResult(nil)
			} else {
				log.Error().Err(err).Msg("handling key share event errored")
				_ = ev.SetResult(err)
			}
		} else {
			_ = ev.SetResult(nil)
		}
	}()

	metricsEpochKGDecryptionTriggersReceived.Inc()
	trigger := ev.Value
	eon, err := ksh.getEonForBlockNumber(ctx, trigger.BlockNumber)
	if err != nil {
		err = errors.Wrap(err, "error retrieving eon for blocknumber")
		return
	}
	log.Debug().Msg("constructing decryption key share")
	keySharesMsg, err := ksh.ConstructDecryptionKeyShares(
		ctx,
		eon,
		trigger.IdentityPreimages,
	)
	if err != nil {
		return
	}

	log.Debug().Msg("sending key share message")
	err = ksh.Messaging.SendMessage(
		ctx,
		keySharesMsg,
		retry.NumberOfRetries(4),
		retry.Interval(200*time.Millisecond),
		retry.LogIdentifier(keySharesMsg.LogInfo()),
	)
	if err != nil {
		err = errors.Wrap(err, "error while sending P2P message")
		return
	}
	metricsEpochKGDecryptionKeySharesSent.Inc()
}

func (ksh *KeyShareHandler) Start(ctx context.Context, group service.Runner) error {
	group.Go(func() error {
		for {
			select {
			case triggerEvent, ok := <-ksh.Trigger:
				if !ok {
					log.Debug().Msg("decryption trigger channel closed, stopping loop")
					return nil
				}
				ksh.handleEvent(ctx, triggerEvent)
			case <-ctx.Done():
				log.Info().Msg("stopping KeyShareHandler due to context cancellation")
				return ctx.Err()
			}
		}
	})
	return nil
}
