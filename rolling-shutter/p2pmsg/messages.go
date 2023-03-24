package p2pmsg

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	anypb "google.golang.org/protobuf/types/known/anypb"

	shcrypto "github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprtopics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/trace"
)

const EnvelopeVersion = "0.0.1"

// All messages to be used in the P2P Gossip have to be included in this slice,
// otherwise they won't be known to the marshaling layer.
var messageTypes = []Message{
	// Keyper messages
	new(DecryptionKey),
	new(DecryptionTrigger),
	new(DecryptionKeyShare),
	new(EonPublicKey),
}

var (
	topicToProtoName = make(map[string]protoreflect.FullName)
	protoNames       = make(map[protoreflect.FullName]bool)
)

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
	protoNames[messageTypeName] = true
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

func Marshal(msg Message, traceContext *TraceContext) ([]byte, error) {
	var msgBytes []byte
	wrappedMsg, err := anypb.New(msg)
	if err != nil {
		return msgBytes, errors.Wrap(err, "failed to wrap protobuf msg in 'any' type")
	}

	envelopedMsg := Envelope{
		Version: EnvelopeVersion,
		Message: wrappedMsg,
		Trace:   traceContext,
	}

	msgBytes, err = proto.Marshal(&envelopedMsg)
	if err != nil {
		return msgBytes, errors.Wrap(err, "failed to marshal p2p message")
	}
	return msgBytes, nil
}

func Unmarshal(data []byte) (Message, *TraceContext, error) {
	envelope := &Envelope{}
	if err := proto.Unmarshal(data, envelope); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal protobuf <Envelope>")
	}

	// Fix the required version for now with an exact version match
	if envelope.GetVersion() != EnvelopeVersion {
		return nil, nil, errors.New("version mismatch")
	}
	traceContext := envelope.GetTrace()

	if !trace.IsEnabled() && traceContext != nil {
		traceContext = nil
	}
	msg, err := envelope.GetMessage().UnmarshalNew()
	if err != nil {
		return nil, traceContext, err
	}
	msgFullName := proto.MessageName(msg)
	if _, ok := protoNames[msgFullName]; !ok {
		return nil, traceContext, errors.Errorf("unknown message type <%s>", msgFullName.Name())
	}

	p2pmess, ok := msg.(Message)
	if !ok {
		return nil, traceContext, errors.Errorf(
			"message of type <%s> does not comply with message interface", msgFullName.Name())
	}
	return p2pmess, traceContext, nil
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
