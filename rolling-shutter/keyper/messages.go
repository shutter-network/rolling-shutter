package keyper

// TODO: This is based on decryptor/messages.go. There's quite some duplication we should get rid
// of, potentially by merging the two and moving them to the p2p package.

import (
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

type message interface {
	implementsMessage()
	GetInstanceID() uint64
}

type (
	decryptionTrigger shmsg.DecryptionTrigger
	eonPublicKey      shmsg.EonPublicKey
)

type decryptionKeyShare struct {
	instanceID  uint64
	epochID     uint64
	keyperIndex uint64
	share       *shcrypto.EpochSecretKeyShare
}

type decryptionKey struct {
	instanceID uint64
	epochID    uint64
	key        *shcrypto.EpochSecretKey
}

func (*decryptionTrigger) implementsMessage()  {}
func (*decryptionKeyShare) implementsMessage() {}
func (*decryptionKey) implementsMessage()      {}
func (*eonPublicKey) implementsMessage()       {}

func (d *decryptionTrigger) GetInstanceID() uint64  { return d.InstanceID }
func (d *decryptionKeyShare) GetInstanceID() uint64 { return d.instanceID }
func (d *decryptionKey) GetInstanceID() uint64      { return d.instanceID }
func (e *eonPublicKey) GetInstanceID() uint64       { return e.InstanceID }

func unmarshalP2PMessage(msg *p2p.Message) (message, error) {
	if msg == nil {
		return nil, nil
	}
	switch msg.Topic {
	case kprtopics.DecryptionTrigger:
		return unmarshalDecryptionTrigger(msg)
	case kprtopics.DecryptionKeyShare:
		return unmarshalDecryptionKeyShare(msg)
	case kprtopics.DecryptionKey:
		return unmarshalDecryptionKey(msg)
	case kprtopics.EonPublicKey:
		return unmarshalEonPublicKey(msg)
	default:
		return nil, errors.New("unhandled topic from P2P message")
	}
}

func unmarshalDecryptionTrigger(msg *p2p.Message) (message, error) {
	decryptionTriggerMsg := shmsg.DecryptionTrigger{}
	if err := proto.Unmarshal(msg.Message, &decryptionTriggerMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption trigger P2P message")
	}
	return (*decryptionTrigger)(&decryptionTriggerMsg), nil
}

func unmarshalDecryptionKeyShare(msg *p2p.Message) (message, error) {
	decryptionKeyShareMsg := shmsg.DecryptionKeyShare{}
	if err := proto.Unmarshal(msg.Message, &decryptionKeyShareMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key share P2P message")
	}

	share := new(shcrypto.EpochSecretKeyShare)
	if err := share.Unmarshal(decryptionKeyShareMsg.Share); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key share P2P message")
	}

	return &decryptionKeyShare{
		instanceID:  decryptionKeyShareMsg.InstanceID,
		epochID:     decryptionKeyShareMsg.EpochID,
		keyperIndex: decryptionKeyShareMsg.KeyperIndex,
		share:       share,
	}, nil
}

func unmarshalDecryptionKey(msg *p2p.Message) (message, error) {
	decryptionKeyMsg := shmsg.DecryptionKey{}
	if err := proto.Unmarshal(msg.Message, &decryptionKeyMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key P2P message")
	}

	key := new(shcrypto.EpochSecretKey)
	if err := key.Unmarshal(decryptionKeyMsg.Key); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key P2P message")
	}

	return &decryptionKey{
		instanceID: decryptionKeyMsg.InstanceID,
		epochID:    decryptionKeyMsg.EpochID,
		key:        key,
	}, nil
}

func unmarshalEonPublicKey(msg *p2p.Message) (message, error) {
	eonKeyMsg := shmsg.EonPublicKey{}
	if err := proto.Unmarshal(msg.Message, &eonKeyMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal eon public key P2P message")
	}
	return (*eonPublicKey)(&eonKeyMsg), nil
}
