package gnosis

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type DecryptionTriggerer struct {
	config                   *configuration.EthnodeConfig
	decryptionTriggerChannel chan *broker.Event[*epochkghandler.DecryptionTrigger]
}

func NewDecryptionTriggerer(
	config *configuration.EthnodeConfig,
	decryptionTriggerChannel chan *broker.Event[*epochkghandler.DecryptionTrigger],
) *DecryptionTriggerer {
	return &DecryptionTriggerer{
		config:                   config,
		decryptionTriggerChannel: decryptionTriggerChannel,
	}
}

func (t *DecryptionTriggerer) Start(ctx context.Context, runner service.Runner) error {
	client, err := ethclient.DialContext(ctx, t.config.EthereumURL)
	if err != nil {
		return err
	}
	runner.Go(func() error {
		headers := make(chan *types.Header)
		headSubscription, err := client.SubscribeNewHead(ctx, headers)
		if err != nil {
			return err
		}

		for {
			select {
			case err := <-headSubscription.Err():
				return err
			case header := <-headers:
				err := t.handleNewHead(ctx, header)
				if err != nil {
					return err
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
	return nil
}

func (t *DecryptionTriggerer) handleNewHead(_ context.Context, header *types.Header) error {
	log.Debug().Uint64("block-number", header.Number.Uint64()).Msg("handling new head")

	// create some random identities for testing
	maxNumIdentityPreimages := uint64(2)
	numIdentityPreimages := header.Number.Uint64()%maxNumIdentityPreimages + 1
	identityPreimages := []identitypreimage.IdentityPreimage{}
	for i := 0; i < int(numIdentityPreimages); i++ {
		n := header.Number.Uint64()*maxNumIdentityPreimages + uint64(i)
		identityPreimage := identitypreimage.Uint64ToIdentityPreimage(n)
		identityPreimages = append(identityPreimages, identityPreimage)
	}
	trigger := epochkghandler.DecryptionTrigger{
		BlockNumber:       header.Number.Uint64(),
		IdentityPreimages: identityPreimages,
	}
	event := broker.NewEvent(&trigger)
	t.decryptionTriggerChannel <- event
	return nil
}
