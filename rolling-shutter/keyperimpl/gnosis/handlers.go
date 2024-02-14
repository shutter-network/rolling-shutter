package gnosis

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type DecryptionKeySharesHandler struct{}

func (h *DecryptionKeySharesHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeyShares{}}
}

func (h *DecryptionKeySharesHandler) ValidateMessage(_ context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	keyShares := msg.(*p2pmsg.DecryptionKeyShares)
	_, ok := keyShares.Extra.(*p2pmsg.DecryptionKeyShares_Gnosis)
	if !ok {
		return pubsub.ValidationReject, errors.Errorf("unexpected extra type %T, expected Gnosis", keyShares.Extra)
	}
	return pubsub.ValidationAccept, nil
}

func (h *DecryptionKeySharesHandler) HandleMessage(_ context.Context, _ p2pmsg.Message) ([]p2pmsg.Message, error) {
	return []p2pmsg.Message{}, nil
}

type DecryptionKeysHandler struct{}

func (h *DecryptionKeysHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeys{}}
}

func (h *DecryptionKeysHandler) ValidateMessage(_ context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	keys := msg.(*p2pmsg.DecryptionKeys)
	_, ok := keys.Extra.(*p2pmsg.DecryptionKeys_Gnosis)
	if !ok {
		return pubsub.ValidationReject, errors.Errorf("unexpected extra type %T, expected Gnosis", keys.Extra)
	}
	return pubsub.ValidationAccept, nil
}

func (h *DecryptionKeysHandler) HandleMessage(_ context.Context, _ p2pmsg.Message) ([]p2pmsg.Message, error) {
	return []p2pmsg.Message{}, nil
}
