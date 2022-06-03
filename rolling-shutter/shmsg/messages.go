package shmsg

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	shcrypto "github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprtopics"
)

// All messages to be used in the P2P Gossip have to be included in this slice,
// otherwise they won't be known to the marshaling layer.
var messageTypes = []P2PMessage{
	// Keyper messages
	new(DecryptionKey),
	new(DecryptionTrigger),
	new(DecryptionKeyShare),
	new(EonPublicKey),
}

var topicToProtoName = make(map[string]protoreflect.FullName)

func init() {
	for _, mess := range messageTypes {
		registerP2PMessage(mess)
	}
}

// Instead of using an envelope for the unmarshalling,
// we simply map one protobuf message type 1 to 1 to a Gossip topic.
func registerP2PMessage(mess P2PMessage) {
	messageTypeName := mess.ProtoReflect().Type().Descriptor().FullName()
	topic := mess.Topic()

	if val, exists := topicToProtoName[topic]; exists {
		if val != messageTypeName {
			err := errors.Errorf("Topic '%s' already has message type <%s> registered. Registering %s failed", topic, val, messageTypeName)
			panic(err)
		}
	}
	topicToProtoName[topic] = messageTypeName
}

// P2PMessage can be send via the p2p protocol.
type P2PMessage interface {
	protoreflect.ProtoMessage
	ImplementsP2PMessage()
	GetInstanceID() uint64
	Topic() string
	LogInfo() string
	Validate() error
}

func NewP2PMessageFromTopic(topic string) (P2PMessage, error) {
	name, ok := topicToProtoName[topic]
	if !ok {
		return nil, errors.Errorf("No message type found for topic <%s>", topic)
	}
	t, err := protoregistry.GlobalTypes.FindMessageByName(name)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while retrieving message type for topic <%s>", topic)
	}
	protomess, ok := t.New().Interface().(P2PMessage)
	if !ok {
		return nil, errors.Errorf("Error while instantiating message type for topic <%s>", topic)
	}
	return protomess, nil
}

func Unmarshal(topic string, data []byte) (P2PMessage, error) {
	var err error

	unmshl, err := NewP2PMessageFromTopic(topic)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to retrieve deserialisation type")
	}

	if err = proto.Unmarshal(data, unmshl); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal protobuf <%s>", reflect.TypeOf(unmshl).String()))
	}

	err = unmshl.Validate()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("verification failed <%s>", reflect.TypeOf(unmshl).String()))
	}
	return unmshl, nil
}

func (*DecryptionTrigger) ImplementsP2PMessage() {
}

func (trigger *DecryptionTrigger) LogInfo() string {
	return fmt.Sprintf("DecryptionTrigger{epochid=%s}", trigger.EpochID)
}

func (*DecryptionTrigger) Topic() string {
	return kprtopics.DecryptionTrigger
}

func (*DecryptionTrigger) Validate() error {
	return nil
}

func (*DecryptionKeyShare) ImplementsP2PMessage() {
}

func (share *DecryptionKeyShare) LogInfo() string {
	return fmt.Sprintf(
		"DecryptionKeyShare{epochid=%s, keyperIndex=%d}",
		share.EpochID,
		share.KeyperIndex,
	)
}

func (*DecryptionKeyShare) Topic() string {
	return kprtopics.DecryptionKeyShare
}

func (share *DecryptionKeyShare) GetEpochSecretKeyShare() (*shcrypto.EpochSecretKeyShare, error) {
	epochSecretKeyShare := new(shcrypto.EpochSecretKeyShare)
	if err := epochSecretKeyShare.Unmarshal(share.GetShare()); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key share P2P message")
	}
	return epochSecretKeyShare, nil
}

func (share *DecryptionKeyShare) Validate() error {
	_, err := share.GetEpochSecretKeyShare()
	return err
}

func (*DecryptionKey) ImplementsP2PMessage() {
}

func (key *DecryptionKey) LogInfo() string {
	return fmt.Sprintf("DecryptionKey{epochid=%s}", key.EpochID)
}

func (*DecryptionKey) Topic() string {
	return kprtopics.DecryptionKey
}

func (key *DecryptionKey) GetEpochSecretKey() (*shcrypto.EpochSecretKey, error) {
	epochSecretKey := new(shcrypto.EpochSecretKey)
	if err := epochSecretKey.Unmarshal(key.GetKey()); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key P2P message")
	}
	return epochSecretKey, nil
}

func (key *DecryptionKey) Validate() error {
	_, err := key.GetEpochSecretKey()
	return err
}

func (*EonPublicKey) ImplementsP2PMessage() {
}

func (e *EonPublicKey) LogInfo() string {
	return fmt.Sprintf(
		"EonPublicKey{activationBlock=%d}",
		e.GetActivationBlock(),
	)
}

func (*EonPublicKey) Topic() string {
	return kprtopics.EonPublicKey
}

func (*EonPublicKey) Validate() error {
	return nil
}

// NewBatchConfig creates a new BatchConfig message.
func NewBatchConfig(
	activationBlockNumber uint64,
	keypers []common.Address,
	threshold uint64,
	keyperConfigIndex uint64,
) *Message {
	var keypersBytes [][]byte
	for _, k := range keypers {
		keypersBytes = append(keypersBytes, k.Bytes())
	}

	return &Message{
		Payload: &Message_BatchConfig{
			BatchConfig: &BatchConfig{
				ActivationBlockNumber: activationBlockNumber,
				Keypers:               keypersBytes,
				Threshold:             threshold,
				KeyperConfigIndex:     keyperConfigIndex,
			},
		},
	}
}

// NewApology creates a new apology message used in the DKG process. This message reveals the
// polyEvals, that where sent encrypted via the PolyEval messages to each accuser.
func NewApology(eon uint64, accusers []common.Address, polyEvals []*big.Int) *Message {
	if len(accusers) != len(polyEvals) {
		panic("bad call to NewApology")
	}

	var accusersBytes [][]byte
	for _, a := range accusers {
		accusersBytes = append(accusersBytes, a.Bytes())
	}

	var polyEvalsBytes [][]byte
	for _, e := range polyEvals {
		polyEvalsBytes = append(polyEvalsBytes, e.Bytes())
	}

	return &Message{
		Payload: &Message_Apology{
			Apology: &Apology{
				Eon:       eon,
				Accusers:  accusersBytes,
				PolyEvals: polyEvalsBytes,
			},
		},
	}
}

func NewAccusation(eon uint64, accused []common.Address) *Message {
	accusedBytes := [][]byte{}
	for _, a := range accused {
		accusedBytes = append(accusedBytes, a.Bytes())
	}
	return &Message{
		Payload: &Message_Accusation{
			Accusation: &Accusation{
				Eon:     eon,
				Accused: accusedBytes,
			},
		},
	}
}

// NewPolyCommitment creates a new poly commitment message containing gamma values.
func NewPolyCommitment(eon uint64, gammas *shcrypto.Gammas) *Message {
	gammaBytes := [][]byte{}
	for _, gamma := range *gammas {
		gammaBytes = append(gammaBytes, gamma.Marshal())
	}

	return &Message{
		Payload: &Message_PolyCommitment{
			PolyCommitment: &PolyCommitment{
				Eon:    eon,
				Gammas: gammaBytes,
			},
		},
	}
}

// NewPolyEval creates a new poly eval message.
func NewPolyEval(eon uint64, receivers []common.Address, encryptedEvals [][]byte) *Message {
	rs := [][]byte{}
	for _, receiver := range receivers {
		rs = append(rs, receiver.Bytes())
	}

	return &Message{
		Payload: &Message_PolyEval{
			PolyEval: &PolyEval{
				Eon:            eon,
				Receivers:      rs,
				EncryptedEvals: encryptedEvals,
			},
		},
	}
}

func NewDKGResult(eon uint64, success bool) *Message {
	return &Message{
		Payload: &Message_DkgResult{
			DkgResult: &DKGResult{
				Eon:     eon,
				Success: success,
			},
		},
	}
}

// NewBlockSeen creates a new BlockSeen message. The keypers send this when they see a new block on
// the main chain that possibly leads to starting a batch config.
func NewBlockSeen(blockNumber uint64) *Message {
	return &Message{
		Payload: &Message_BlockSeen{
			BlockSeen: &BlockSeen{
				BlockNumber: blockNumber,
			},
		},
	}
}

// NewCheckIn creates a new CheckIn message.
func NewCheckIn(validatorPublicKey []byte, encryptionKey *ecies.PublicKey) *Message {
	encryptionKeyECDSA := encryptionKey.ExportECDSA()
	return &Message{
		Payload: &Message_CheckIn{
			CheckIn: &CheckIn{
				ValidatorPublicKey:  validatorPublicKey,
				EncryptionPublicKey: crypto.CompressPubkey(encryptionKeyECDSA),
			},
		},
	}
}
