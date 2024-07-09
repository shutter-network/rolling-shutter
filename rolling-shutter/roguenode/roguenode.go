package roguenode

import (
	"context"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type RogueNode struct {
	config    *Config
	messaging p2p.Messaging
}

func New(config *Config) *RogueNode {
	return &RogueNode{
		config: config,
	}
}

func (node *RogueNode) Start(ctx context.Context, runner service.Runner) error {
	incomingMessageCh := make(chan p2pmsg.Message)
	messageSender, err := p2p.New(node.config.P2P)
	if err != nil {
		return errors.Wrap(err, "failed to initialize p2p messaging")
	}
	messageSender.AddMessageHandler(NewHandler(incomingMessageCh))
	node.messaging = messageSender

	runner.Go(func() error {
		return node.sendMessages(ctx, incomingMessageCh)
	})
	return runner.StartService(messageSender)
}

type Handler struct {
	incomingMessageCh chan p2pmsg.Message
	lastMessageTime   time.Time
}

func NewHandler(incomingMessageCh chan p2pmsg.Message) *Handler {
	return &Handler{
		incomingMessageCh: incomingMessageCh,
	}
}

func (*Handler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeys{}}
}

func (handler *Handler) ValidateMessage(_ context.Context, _ p2pmsg.Message) (pubsub.ValidationResult, error) {
	return pubsub.ValidationAccept, nil
}

func (handler *Handler) HandleMessage(_ context.Context, msg p2pmsg.Message) ([]p2pmsg.Message, error) {
	handler.incomingMessageCh <- msg
	return []p2pmsg.Message{}, nil
}
