package p2pmsg

import (
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	anypb "google.golang.org/protobuf/types/known/anypb"

	shcrypto "github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprtopics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/trace"
)

const EnvelopeVersion = "0.0.1"

// Message can be send via the p2p protocol.
type Message interface {
	protoreflect.ProtoMessage
	GetInstanceID() uint64
	Topic() string
	LogInfo() string
	Validate() error
	String() string
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
	var traceContext *TraceContext
	if trace.IsEnabled() {
		traceContext = envelope.GetTrace()
	}

	msg, err := envelope.GetMessage().UnmarshalNew()
	if err != nil {
		return nil, traceContext, err
	}

	p2pmess, ok := msg.(Message)
	if !ok {
		return nil, traceContext, errors.Errorf(
			"message of type <%s> does not comply with message interface", proto.MessageName(msg))
	}
	return p2pmess, traceContext, nil
}

func (trigger *DecryptionTrigger) LogInfo() string {
	epochID, _ := epochid.BytesToEpochID(trigger.EpochID)
	return fmt.Sprintf("DecryptionTrigger{epochid=%x}", epochID.String())
}

func (*DecryptionTrigger) Topic() string {
	return kprtopics.DecryptionTrigger
}

func (*DecryptionTrigger) Validate() error {
	return nil
}

func (share *DecryptionKeyShares) LogInfo() string {
	return fmt.Sprintf(
		"DecryptionKeyShares{keyperIndex=%d}",
		share.KeyperIndex,
	)
}

func (*DecryptionKeyShares) Topic() string {
	return kprtopics.DecryptionKeyShares
}

func (share *KeyShare) GetEpochSecretKeyShare() (*shcrypto.EpochSecretKeyShare, error) {
	epochSecretKeyShare := new(shcrypto.EpochSecretKeyShare)
	if err := epochSecretKeyShare.Unmarshal(share.GetShare()); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal decryption key share P2P message")
	}
	return epochSecretKeyShare, nil
}

func (share *DecryptionKeyShares) Validate() error {
	for _, sh := range share.GetShares() {
		_, err := sh.GetEpochSecretKeyShare()
		if err != nil {
			return err
		}
	}
	return nil
}

func (key *DecryptionKey) LogInfo() string {
	return fmt.Sprintf("DecryptionKey{epochid=%x}", key.EpochID)
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
