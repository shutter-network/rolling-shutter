package p2pnode

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

// dummyMessageHandler validates all p2p messages and emits a log message for each p2p message.
type dummyMessageHandler struct{}

func (dummyMessageHandler) ValidateMessage(_ context.Context, _ p2pmsg.Message) (pubsub.ValidationResult, error) {
	return pubsub.ValidationAccept, nil
}

func (dummyMessageHandler) HandleMessage(
	_ context.Context,
	msg p2pmsg.Message,
) ([]p2pmsg.Message, error) {
	log.Info().Str("message", msg.String()).Msg("received message")
	return nil, nil
}

func (dummyMessageHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{
		&p2pmsg.DecryptionKeyShares{},
		&p2pmsg.DecryptionKeys{},
		&p2pmsg.DecryptionTrigger{},
		&p2pmsg.EonPublicKey{},
		&p2pmsg.Commitment{},
	}
}

type P2PNode struct {
	config *Config
}

func New(config *Config) *P2PNode {
	return &P2PNode{
		config: config,
	}
}

func (node *P2PNode) Start(_ context.Context, runner service.Runner) error {
	services := []service.Service{}

	p2pHandler, err := p2p.New(node.config.P2P)
	if err != nil {
		if err != nil {
			return errors.Wrap(err, "failed to initialize p2p messaging")
		}
	}
	if node.config.ListenMessages {
		p2pHandler.AddMessageHandler(dummyMessageHandler{})
	}
	services = append(services, p2pHandler)

	if node.config.Metrics.Enabled {
		metricsServer := metricsserver.New(node.config.Metrics)
		services = append(services, metricsServer)
	}
	return runner.StartService(services...)
}
