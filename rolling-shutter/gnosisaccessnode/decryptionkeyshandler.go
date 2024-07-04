package gnosisaccessnode

import (
	"bytes"
	"context"
	"math"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	obskeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
)

type DecryptionKeysHandler struct {
	config  *Config
	storage *Storage
}

func NewDecryptionKeysHandler(config *Config, storage *Storage) *DecryptionKeysHandler {
	return &DecryptionKeysHandler{
		config:  config,
		storage: storage,
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

func (handler *DecryptionKeysHandler) validateCommonFields(keys *p2pmsg.DecryptionKeys) (pubsub.ValidationResult, error) {
	if keys.InstanceID != handler.config.InstanceID {
		return pubsub.ValidationReject,
			errors.Errorf("instance ID mismatch (want=%d, have=%d)", handler.config.InstanceID, keys.GetInstanceID())
	}
	if keys.Eon > math.MaxInt64 {
		return pubsub.ValidationReject, errors.Errorf("eon %d overflows int64", keys.Eon)
	}

	if len(keys.Keys) == 0 {
		return pubsub.ValidationReject, errors.New("no keys in message")
	}
	if len(keys.Keys) > int(handler.config.MaxNumKeysPerMessage) {
		return pubsub.ValidationReject, errors.Errorf(
			"too many keys in message (%d > %d)",
			len(keys.Keys),
			handler.config.MaxNumKeysPerMessage,
		)
	}

	eonKey, ok := handler.storage.GetEonKey(keys.Eon)
	if !ok {
		return pubsub.ValidationReject, errors.Errorf("no eon key found for eon %d", keys.Eon)
	}

	for i, k := range keys.Keys {
		epochSecretKey, err := k.GetEpochSecretKey()
		if err != nil {
			return pubsub.ValidationReject, err
		}
		ok, err := shcrypto.VerifyEpochSecretKey(epochSecretKey, eonKey, k.Identity)
		if err != nil {
			return pubsub.ValidationReject, errors.Wrapf(err, "error while checking epoch secret key for identity %x", k.Identity)
		}
		if !ok {
			return pubsub.ValidationReject, errors.Errorf("epoch secret key for identity %x is not valid", k.Identity)
		}

		if i > 0 && bytes.Compare(k.Identity, keys.Keys[i-1].Identity) < 0 {
			return pubsub.ValidationReject, errors.Errorf("keys not ordered")
		}
	}

	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionKeysHandler) validateGnosisFields(keys *p2pmsg.DecryptionKeys) (pubsub.ValidationResult, error) {
	res, err := gnosis.ValidateDecryptionKeysBasic(keys)
	if res != pubsub.ValidationAccept || err != nil {
		return res, err
	}
	extra := keys.Extra.(*p2pmsg.DecryptionKeys_Gnosis).Gnosis

	// TODO: populate this from the contract
	keyperSet := &obskeyperdatabase.KeyperSet{
		KeyperConfigIndex:     0,
		ActivationBlockNumber: 0,
		Keypers:               []string{},
		Threshold:             0,
	}

	res, err = gnosis.ValidateDecryptionKeysSignatures(keys, extra, keyperSet)
	if res != pubsub.ValidationAccept || err != nil {
		return res, err
	}

	return pubsub.ValidationAccept, nil
}

func (handler *DecryptionKeysHandler) HandleMessage(_ context.Context, _ p2pmsg.Message) ([]p2pmsg.Message, error) {
	return nil, nil
}
