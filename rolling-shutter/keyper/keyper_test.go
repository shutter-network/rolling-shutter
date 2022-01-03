package keyper

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"google.golang.org/protobuf/proto"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shuttermint/keyper/kprdb"
	"github.com/shutter-network/shutter/shuttermint/keyper/kprtopics"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
	"github.com/shutter-network/shutter/shuttermint/medley/testdb"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func TestDecryptionKeyValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	config := newTestConfig(t)
	keyperIndex := uint64(1)
	epochID := uint64(0)
	wrongEpochID := uint64(1)
	tkg := initializeEon(ctx, t, db, config, keyperIndex)
	secretKey := tkg.EpochSecretKey(epochID).Marshal()
	keyshare := tkg.EpochSecretKeyShare(epochID, keyperIndex).Marshal()

	kpr := keyper{config: config, db: db}
	var peerID peer.ID

	tests := []struct {
		name      string
		validator pubsub.Validator
		valid     bool
		msg       shmsg.P2PMessage
	}{
		{
			name:      "valid decryption key",
			validator: kpr.validateDecryptionKey,
			valid:     true,
			msg: &shmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				EpochID:    epochID,
				Key:        secretKey,
			},
		},
		{
			name:      "invalid decryption key wrong epoch",
			validator: kpr.validateDecryptionKey,
			valid:     false,
			msg: &shmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				EpochID:    wrongEpochID,
				Key:        secretKey,
			},
		},
		{
			name:      "invalid decryption key wrong instance ID",
			validator: kpr.validateDecryptionKey,
			valid:     false,
			msg: &shmsg.DecryptionKey{
				InstanceID: config.InstanceID + 1,
				EpochID:    epochID,
				Key:        secretKey,
			},
		},
		{
			name:      "valid decryption key share",
			validator: kpr.validateDecryptionKeyShare,
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
			validator: kpr.validateDecryptionKeyShare,
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
			validator: kpr.validateDecryptionKeyShare,
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
			validator: kpr.validateDecryptionKeyShare,
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

func TestTriggerValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	config := newTestConfig(t)
	kpr := keyper{config: config, db: db}

	collatorKey1, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	collatorAddress1 := ethcrypto.PubkeyToAddress(collatorKey1.PublicKey)

	collatorKey2, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	collatorAddress2 := ethcrypto.PubkeyToAddress(collatorKey2.PublicKey)

	// Make a db with collator 1 from a certain block and collator 2 afterwards
	activationBlk1 := int64(0)
	epochID1 := uint64(0)
	activationBlk2 := int64(123)
	epochID2 := epochid.New(0, uint32(activationBlk2))
	assert.NilError(t, err)
	collator1 := shdb.EncodeAddress(collatorAddress1)
	collator2 := shdb.EncodeAddress(collatorAddress2)
	err = db.InsertChainCollator(ctx, kprdb.InsertChainCollatorParams{
		ActivationBlockNumber: activationBlk1,
		Collator:              collator1,
	})
	assert.NilError(t, err)
	err = db.InsertChainCollator(ctx, kprdb.InsertChainCollatorParams{
		ActivationBlockNumber: activationBlk2,
		Collator:              collator2,
	})
	assert.NilError(t, err)

	var peerID peer.ID

	tests := []struct {
		name       string
		valid      bool
		instanceID uint64
		epochID    uint64
		privKey    *ecdsa.PrivateKey
	}{
		{
			name:       "valid trigger collator 1",
			valid:      true,
			instanceID: config.InstanceID,
			epochID:    epochID1,
			privKey:    collatorKey1,
		},
		{
			name:       "valid trigger collator 2",
			valid:      true,
			instanceID: config.InstanceID,
			epochID:    epochID2,
			privKey:    collatorKey2,
		},
		{
			name:       "invalid trigger wrong collator 1",
			valid:      false,
			instanceID: config.InstanceID,
			epochID:    epochID2,
			privKey:    collatorKey1,
		},
		{
			name:       "invalid trigger wrong collator 2",
			valid:      false,
			instanceID: config.InstanceID,
			epochID:    epochID1,
			privKey:    collatorKey2,
		},
		{
			name:       "invalid trigger wrong instanceID",
			valid:      false,
			instanceID: config.InstanceID + 1,
			epochID:    epochID1,
			privKey:    collatorKey1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := shmsg.NewSignedDecryptionTrigger(
				tc.instanceID,
				tc.epochID,
				[][]byte{},
				tc.privKey,
			)
			assert.NilError(t, err)
			pubsubMsg, err := makePubSubMessage(msg, kprtopics.DecryptionTrigger)
			if err != nil {
				t.Fatalf("Error in makePubSubMessage: %s", err)
			}
			assert.Equal(t, kpr.validateDecryptionTrigger(ctx, peerID, pubsubMsg), tc.valid,
				"validate failed valid=%t", tc.valid)
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
