package keyper

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"google.golang.org/protobuf/proto"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func TestDecryptionKeyValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewKeyperTestDB(ctx, t)
	defer closedb()

	config := newTestConfig(t)
	keyperIndex := uint64(1)
	epochID := uint64(0)
	wrongEpochID := uint64(1)
	tkg := initializeEon(ctx, t, db, config, keyperIndex)
	secretKey, _ := tkg.EpochSecretKey(epochID).GobEncode()

	kpr := keyper{config: config}
	validator := kpr.makeDecryptionKeyValidator(db)
	var peerID peer.ID

	tests := []struct {
		name  string
		valid bool
		msg   shmsg.P2PMessage
	}{
		{
			name:  "valid decryption key",
			valid: true,
			msg: &shmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				EpochID:    epochID,
				Key:        secretKey,
			},
		},
		{
			name:  "invalid decryption key wrong epoch",
			valid: false,
			msg: &shmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				EpochID:    wrongEpochID,
				Key:        secretKey,
			},
		},
		{
			name:  "invalid decryption key wrong instance ID",
			valid: false,
			msg: &shmsg.DecryptionKey{
				InstanceID: config.InstanceID + 1,
				EpochID:    epochID,
				Key:        secretKey,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pubsubMessage, err := makePubSubMessage(tc.msg, tc.msg.Topic())
			if err != nil {
				t.Fatalf("Error in makePubSubMessage: %s", err)
			}
			assert.Equal(t, validator(ctx, peerID, pubsubMessage), tc.valid,
				"validate failed valid=%t msg=%+v type=%T", tc.valid, tc.msg, tc.msg)
		})
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
