package decryptor

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func TestMessageValidators(t *testing.T) {
	ctx := context.Background()
	var peerID peer.ID
	validID := uint64(1)
	wrongID := uint64(0)
	cfg := Config{
		InstanceID: validID,
	}
	d := New(cfg)

	validMessage := shmsg.DecryptionKey{
		InstanceID: validID,
	}
	wrongMessage := shmsg.DecryptionKey{
		InstanceID: wrongID,
	}

	validators := d.makeMessagesValidators()
	var emptyValidator pubsub.Validator

	for topic, validator := range validators {
		assert.Equal(t, reflect.TypeOf(validator), reflect.TypeOf(emptyValidator))

		validPubsubMessage, err := makePubSubMessage(&validMessage, topic)
		if err != nil {
			t.Fatalf("Error while making valid message")
		}

		wrongPubsubMessage, err := makePubSubMessage(&wrongMessage, topic)
		if err != nil {
			t.Fatalf("Error while making valid message")
		}

		assert.Check(t, validator(ctx, peerID, validPubsubMessage))
		assert.Equal(t, validator(ctx, peerID, wrongPubsubMessage), false)
	}
}

// makePubSubMessage makes a pubsub.Message corresponding to the type received by gossip validators.
func makePubSubMessage(message shmsg.P2PMessage, topic string) (*pubsub.Message, error) {
	var messageBytes []byte
	var err error

	switch m := message.(type) {
	case *shmsg.DecryptionKey:
		messageBytes, err = proto.Marshal(m)
	case *shmsg.CipherBatch:
		messageBytes, err = proto.Marshal(m)
	case *shmsg.AggregatedDecryptionSignature:
		messageBytes, err = proto.Marshal(m)
	default:
		return nil, errors.Errorf("received message of unexpected type: %s", message)
	}
	if err != nil {
		return nil, err
	}

	P2PMessage := p2p.Message{
		Topic:    topic,
		Message:  messageBytes,
		SenderID: "",
	}
	b, err := json.Marshal(&P2PMessage)
	if err != nil {
		return nil, err
	}

	pubsubMessage := pubsub.Message{
		Message: &pb.Message{
			Data: b,
		},
		ReceivedFrom:  "",
		ValidatorData: nil,
	}

	return &pubsubMessage, nil
}
