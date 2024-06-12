package p2pnode

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/rs/zerolog/log"

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
	}
}

func New(config *Config) (service.Service, error) {
	p2pHandler, err := p2p.New(config.P2P)
	if err != nil {
		return nil, err
	}
	if config.ListenMessages {
		p2pHandler.AddMessageHandler(dummyMessageHandler{})
	}
	return p2pHandler, nil
}
