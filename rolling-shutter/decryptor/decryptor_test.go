package decryptor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func TestMessageValidators(t *testing.T) {
	ctx := context.Background()
	var peerID peer.ID
	d := New(Config{
		InstanceID: 123,
	})
	validators := d.makeMessagesValidators()
	validDecryptionKey := make([]byte, 64)
	validSignature := make([]byte, 128)
	validHash := make([]byte, 64)
	tests := []struct {
		valid bool
		msg   shmsg.P2PMessage
	}{
		{
			valid: true,
			msg: &shmsg.DecryptionKey{
				InstanceID: d.Config.InstanceID,
				Key:        validDecryptionKey,
			},
		},
		{
			valid: true,
			msg: &shmsg.AggregatedDecryptionSignature{
				InstanceID:          d.Config.InstanceID,
				AggregatedSignature: validSignature,
				SignedHash:          validHash,
			},
		},
		{
			valid: true,
			msg: &shmsg.CipherBatch{
				InstanceID: d.Config.InstanceID,
			},
		},
		{
			valid: false,
			msg: &shmsg.DecryptionKey{
				InstanceID: d.Config.InstanceID + 1,
			},
		},
		{
			valid: false,
			msg: &shmsg.AggregatedDecryptionSignature{
				InstanceID: d.Config.InstanceID - 1,
			},
		},
		{
			valid: false,
			msg: &shmsg.CipherBatch{
				InstanceID: d.Config.InstanceID + 2,
			},
		},
	}
	for _, tc := range tests {
		pubsubMessage, err := makePubSubMessage(tc.msg, tc.msg.Topic())
		if err != nil {
			t.Fatalf("Error in makePubSubMessage: %s", err)
		}
		validate := validators[pubsubMessage.GetTopic()]
		assert.Assert(t, validate != nil)
		assert.Equal(t, validate(ctx, peerID, pubsubMessage), tc.valid,
			"validate failed valid=%t msg=%+v type=%T", tc.valid, tc.msg, tc.msg)
	}
}

// makePubSubMessage makes a pubsub.Message corresponding to the type received by gossip validators.
func makePubSubMessage(message shmsg.P2PMessage, topic string) (*pubsub.Message, error) {
	messageBytes, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(&p2p.Message{
		Topic:    topic,
		Message:  messageBytes,
		SenderID: "",
	})
	if err != nil {
		return nil, err
	}

	pubsubMessage := pubsub.Message{
		Message: &pb.Message{
			Data:  b,
			Topic: &topic,
		},
		ReceivedFrom:  "",
		ValidatorData: nil,
	}

	return &pubsubMessage, nil
}
