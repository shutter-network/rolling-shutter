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

func unmarshalP2PMessage(msg *p2p.Message) (shmsg.P2PMessage, error) {
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

func unmarshalDecryptionTrigger(msg *p2p.Message) (*shmsg.DecryptionTrigger, error) {
	decryptionTriggerMsg := shmsg.DecryptionTrigger{}
	if err := proto.Unmarshal(msg.Message, &decryptionTriggerMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption trigger P2P message")
	}
	return &decryptionTriggerMsg, nil
}

func unmarshalDecryptionKeyShare(msg *p2p.Message) (*shmsg.DecryptionKeyShare, error) {
	decryptionKeyShareMsg := shmsg.DecryptionKeyShare{}
	if err := proto.Unmarshal(msg.Message, &decryptionKeyShareMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key share P2P message")
	}

	share := new(shcrypto.EpochSecretKeyShare)
	if err := share.Unmarshal(decryptionKeyShareMsg.Share); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key share P2P message")
	}
	return &decryptionKeyShareMsg, nil
}

func unmarshalDecryptionKey(msg *p2p.Message) (*shmsg.DecryptionKey, error) {
	decryptionKeyMsg := shmsg.DecryptionKey{}
	if err := proto.Unmarshal(msg.Message, &decryptionKeyMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key P2P message")
	}

	key := new(shcrypto.EpochSecretKey)
	if err := key.Unmarshal(decryptionKeyMsg.Key); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key P2P message")
	}

	return &decryptionKeyMsg, nil
}

func unmarshalEonPublicKey(msg *p2p.Message) (*shmsg.EonPublicKey, error) {
	eonKeyMsg := shmsg.EonPublicKey{}
	if err := proto.Unmarshal(msg.Message, &eonKeyMsg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal eon public key P2P message")
	}
	return &eonKeyMsg, nil
}
