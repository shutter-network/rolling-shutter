package p2pmsg

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	shcrypto "github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprtopics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
)

// All messages to be used in the P2P Gossip have to be included in this slice,
// otherwise they won't be known to the marshaling layer.
var messageTypes = []Message{
	// Keyper messages
	new(DecryptionKey),
	new(DecryptionTrigger),
	new(DecryptionKeyShare),
	new(EonPublicKey),
}

var topicToProtoName = make(map[string]protoreflect.FullName)

func init() {
	for _, mess := range messageTypes {
		registerMessage(mess)
	}
}

// Instead of using an envelope for the unmarshalling,
// we simply map one protobuf message type 1 to 1 to a Gossip topic.
func registerMessage(mess Message) {
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

// Message can be send via the p2p protocol.
type Message interface {
	protoreflect.ProtoMessage
	ImplementsP2PMessage()
	GetInstanceID() uint64
	Topic() string
	LogInfo() string
	Validate() error
}

func NewMessageFromTopic(topic string) (Message, error) {
	name, ok := topicToProtoName[topic]
	if !ok {
		return nil, errors.Errorf("No message type found for topic <%s>", topic)
	}
	t, err := protoregistry.GlobalTypes.FindMessageByName(name)
	if err != nil {
		return nil, errors.Wrapf(err, "Error while retrieving message type for topic <%s>", topic)
	}
	protomess, ok := t.New().Interface().(Message)
	if !ok {
		return nil, errors.Errorf("Error while instantiating message type for topic <%s>", topic)
	}
	return protomess, nil
}

func Unmarshal(topic string, data []byte) (Message, error) {
	var err error

	unmshl, err := NewMessageFromTopic(topic)
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
	epochID, _ := epochid.BytesToEpochID(trigger.EpochID)
	return fmt.Sprintf("DecryptionTrigger{epochid=%s}", epochID.String())
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
