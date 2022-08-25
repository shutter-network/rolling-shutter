package decryptor

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrtopics"
	"github.com/shutter-network/shutter/shuttermint/medley/bitfield"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

type decryptionSignature struct {
	instanceID uint64
	epochID    []byte
	signedHash common.Hash
	signature  *shbls.Signature
	signers    bitfield.Bitfield
}

type aggregatedDecryptionSignature struct {
	instanceID uint64
	epochID    []byte
	signedHash common.Hash
	signature  *shbls.Signature
	signers    bitfield.Bitfield
}

type (
	cipherBatch   shmsg.CipherBatch
	decryptionKey shmsg.DecryptionKey
)

type message interface {
	implementsMessage()
	GetInstanceID() uint64
}

func (*decryptionSignature) implementsMessage()           {}
func (*aggregatedDecryptionSignature) implementsMessage() {}
func (*decryptionKey) implementsMessage()                 {}
func (*cipherBatch) implementsMessage()                   {}

func (d *decryptionSignature) GetInstanceID() uint64           { return d.instanceID }
func (d *aggregatedDecryptionSignature) GetInstanceID() uint64 { return d.instanceID }
func (d *decryptionKey) GetInstanceID() uint64                 { return d.InstanceID }
func (c *cipherBatch) GetInstanceID() uint64                   { return c.DecryptionTrigger.InstanceID }

func unmarshalP2PMessage(msg *p2p.Message) (message, error) {
	if msg == nil {
		return nil, nil
	}
	switch msg.Topic {
	case dcrtopics.DecryptionKey:
		decryptionKeyMsg := shmsg.DecryptionKey{}
		if err := proto.Unmarshal(msg.Message, &decryptionKeyMsg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal decryption key P2P message")
		}

		key := new(shcrypto.EpochSecretKey)
		if err := key.Unmarshal(decryptionKeyMsg.Key); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal decryption key")
		}
		encodedKey, err := key.GobEncode()
		if err != nil {
			return nil, errors.Wrap(err, "failed to encode decryption key")
		}

		return &decryptionKey{
			InstanceID: decryptionKeyMsg.InstanceID,
			EpochID:    decryptionKeyMsg.EpochID,
			Key:        encodedKey,
		}, nil

	case dcrtopics.CipherBatch:
		cipherBatchMsg := shmsg.CipherBatch{}
		if err := proto.Unmarshal(msg.Message, &cipherBatchMsg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal cipher batch P2P message")
		}
		return (*cipherBatch)(&cipherBatchMsg), nil

	case dcrtopics.DecryptionSignature:
		decryptionSignatureMsg := shmsg.DecryptionSignature{}
		if err := proto.Unmarshal(msg.Message, &decryptionSignatureMsg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal decryption signature P2P message")
		}
		signature := new(shbls.Signature)
		if err := signature.Unmarshal(decryptionSignatureMsg.Signature); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal decryption signature")
		}
		return &decryptionSignature{
			instanceID: decryptionSignatureMsg.InstanceID,
			epochID:    decryptionSignatureMsg.EpochID,
			signedHash: common.BytesToHash(decryptionSignatureMsg.SignedHash),
			signature:  signature,
			signers:    bitfield.Bitfield(decryptionSignatureMsg.SignerBitfield),
		}, nil
	case dcrtopics.AggregatedDecryptionSignature:
		decryptionSignatureMsg := shmsg.AggregatedDecryptionSignature{}
		if err := proto.Unmarshal(msg.Message, &decryptionSignatureMsg); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal aggregated decryption signature P2P message")
		}
		signature := new(shbls.Signature)
		if err := signature.Unmarshal(decryptionSignatureMsg.AggregatedSignature); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal decryption signature")
		}
		return &aggregatedDecryptionSignature{
			instanceID: decryptionSignatureMsg.InstanceID,
			epochID:    decryptionSignatureMsg.EpochID,
			signedHash: common.BytesToHash(decryptionSignatureMsg.SignedHash),
			signature:  signature,
			signers:    bitfield.Bitfield(decryptionSignatureMsg.SignerBitfield),
		}, nil

	default:
		return nil, &unhandledTopicError{msg.Topic, "unhandled topic from P2P message"}
	}
}

type unhandledTopicError struct {
	topic string
	msg   string
}

func (e *unhandledTopicError) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.topic)
}
