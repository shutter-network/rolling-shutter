package gnosisaccessnode

import (
	"context"
	"math"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

const (
	// TODO: pull these from keyper set manager
	n         = 7
	threshold = 5
)

type DecryptionKeysHandler struct {
	config *Config
}

func NewDecryptionKeysHandler(config *Config) *DecryptionKeysHandler {
	return &DecryptionKeysHandler{
		config: config,
	}
}

func (*DecryptionKeysHandler) MessagePrototypes() []p2pmsg.Message {
	return []p2pmsg.Message{&p2pmsg.DecryptionKeys{}}
}

func (handler *DecryptionKeysHandler) ValidateMessage(_ context.Context, msg p2pmsg.Message) (pubsub.ValidationResult, error) {
	key := msg.(*p2pmsg.DecryptionKeys)
	result, err := handler.validateCommonFields(key)
	if result != pubsub.ValidationAccept || err != nil {
		return result, err
	}
	result, err = handler.validateGnosisFields(key)
	if result != pubsub.ValidationAccept || err != nil {
		return result, err
	}
	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionKeysHandler) validateCommonFields(key *p2pmsg.DecryptionKeys) (pubsub.ValidationResult, error) {
	if key.InstanceID != handler.config.InstanceID {
		return pubsub.ValidationReject,
			errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.InstanceID, key.GetInstanceID())
	}
	if key.Eon > math.MaxInt64 {
		return pubsub.ValidationReject, errors.Errorf("eon %d overflows int64", key.Eon)
	}

	if len(key.Keys) == 0 {
		return pubsub.ValidationReject, errors.New("no keys in message")
	}
	if len(key.Keys) > int(handler.config.MaxNumKeysPerMessage) {
		return pubsub.ValidationReject, errors.Errorf(
			"too many keys in message (%d > %d)",
			len(key.Keys),
			handler.config.MaxNumKeysPerMessage,
		)
	}
	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionKeysHandler) validateGnosisFields(keys *p2pmsg.DecryptionKeys) (pubsub.ValidationResult, error) {
	extraWrapped, ok := keys.Extra.(*p2pmsg.DecryptionKeys_Gnosis)
	if !ok {
		return pubsub.ValidationReject, errors.Errorf("unexpected extra type %T, expected Gnosis", keys.Extra)
	}
	extra := extraWrapped.Gnosis
	if extra == nil {
		return pubsub.ValidationReject, errors.New("missing extra Gnosis data")
	}

	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionKeysHandler) HandleMessage(_ context.Context, _ p2pmsg.Message) ([]p2pmsg.Message, error) {
	return nil, nil
}
