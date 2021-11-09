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
	secretKey := tkg.EpochSecretKey(epochID).Marshal()
	keyshare := tkg.EpochSecretKeyShare(epochID, keyperIndex).Marshal()

	kpr := keyper{config: config}
	keyValidator := kpr.makeDecryptionKeyValidator(db)
	keyShareValidator := kpr.makeKeyShareValidator(db)
	var peerID peer.ID

	tests := []struct {
		name      string
		validator pubsub.Validator
		valid     bool
		msg       shmsg.P2PMessage
	}{
		{
			name:      "valid decryption key",
			validator: keyValidator,
			valid:     true,
			msg: &shmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				EpochID:    epochID,
				Key:        secretKey,
			},
		},
		{
			name:      "invalid decryption key wrong epoch",
			validator: keyValidator,
			valid:     false,
			msg: &shmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				EpochID:    wrongEpochID,
				Key:        secretKey,
			},
		},
		{
			name:      "invalid decryption key wrong instance ID",
			validator: keyValidator,
			valid:     false,
			msg: &shmsg.DecryptionKey{
				InstanceID: config.InstanceID + 1,
				EpochID:    epochID,
				Key:        secretKey,
			},
		},
		{
			name:      "valid decryption key share",
			validator: keyShareValidator,
			valid:     true,
			msg: &shmsg.DecryptionKeyShare{
				InstanceID:  config.InstanceID,
				EpochID:     epochID,
				KeyperIndex: keyperIndex,
				Share:       keyshare,
			},
		},
		{
			name:      "invalid decryption key share wrong epoch",
			validator: keyShareValidator,
			valid:     false,
			msg: &shmsg.DecryptionKeyShare{
				InstanceID:  config.InstanceID,
				EpochID:     epochID + 1,
				KeyperIndex: keyperIndex,
				Share:       keyshare,
			},
		},
		{
			name:      "invalid decryption key share wrong instance ID",
			validator: keyShareValidator,
			valid:     false,
			msg: &shmsg.DecryptionKeyShare{
				InstanceID:  config.InstanceID + 1,
				EpochID:     epochID,
				KeyperIndex: keyperIndex,
				Share:       keyshare,
			},
		},
		{
			name:      "invalid decryption key share wrong keyper index",
			validator: keyShareValidator,
			valid:     false,
			msg: &shmsg.DecryptionKeyShare{
				InstanceID:  config.InstanceID,
				EpochID:     epochID,
				KeyperIndex: keyperIndex + 1,
				Share:       keyshare,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pubsubMessage, err := makePubSubMessage(tc.msg, tc.msg.Topic())
			if err != nil {
				t.Fatalf("Error in makePubSubMessage: %s", err)
			}
			assert.Equal(t, tc.validator(ctx, peerID, pubsubMessage), tc.valid,
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
