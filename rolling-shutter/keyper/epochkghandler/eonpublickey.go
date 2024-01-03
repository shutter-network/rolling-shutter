package epochkghandler

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

func NewEonPublicKeyHandler(config Config, _ *pgxpool.Pool) p2p.MessageHandler {
	return &EonPublicKeyHandler{config: config}
}

type EonPublicKeyHandler struct {
	config Config
}

func (*EonPublicKeyHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.EonPublicKey{}}
}

func (handler *EonPublicKeyHandler) ValidateMessage(_ context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	key := msg.(*p2pmsg.EonPublicKey)
	if key.GetInstanceID() != handler.config.GetInstanceID() {
		return pubsub.ValidationReject, errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.GetInstanceID(), key.GetInstanceID())
	}
	return pubsub.ValidationAccept, nil
}

func (handler *EonPublicKeyHandler) HandleMessage(context.Context, p2pmsg.Message) ([]p2pmsg.Message, error) {
	return nil, nil
}
