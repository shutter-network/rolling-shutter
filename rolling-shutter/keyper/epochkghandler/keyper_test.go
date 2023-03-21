package epochkghandler

import (
	"context"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pmsg"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func TestDecryptionKeyValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	keyperIndex := uint64(1)
	eon := uint64(0)
	epochID, _ := epochid.BigToEpochID(common.Big0)
	wrongEpochID, _ := epochid.BigToEpochID(common.Big1)
	tkg := initializeEon(ctx, t, db, keyperIndex)
	secretKey := tkg.EpochSecretKey(epochID).Marshal()
	keyshare := tkg.EpochSecretKeyShare(epochID, keyperIndex).Marshal()

	p2pHandler := p2p.New(p2p.Config{})

	kpr := New(config.Address, config.InstanceID, dbpool)

	var peerID peer.ID

	validateDecryptionKey := p2p.AddValidator(p2pHandler, kpr.validateDecryptionKey)
	validateDecryptionKeyShare := p2p.AddValidator(p2pHandler, kpr.validateDecryptionKeyShare)

	tests := []struct {
		name      string
		validator pubsub.Validator
		valid     bool
		msg       p2pmsg.Message
	}{
		{
			name:      "valid decryption key",
			valid:     true,
			validator: validateDecryptionKey,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				Eon:        eon,
				EpochID:    epochID.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:      "invalid decryption key wrong epoch",
			valid:     false,
			validator: validateDecryptionKey,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.InstanceID,
				Eon:        eon,
				EpochID:    wrongEpochID.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:      "invalid decryption key wrong instance ID",
			valid:     false,
			validator: validateDecryptionKey,
			msg: &p2pmsg.DecryptionKey{
				InstanceID: config.InstanceID + 1,
				Eon:        eon,
				EpochID:    epochID.Bytes(),
				Key:        secretKey,
			},
		},
		{
			name:      "valid decryption key share",
			valid:     true,
			validator: validateDecryptionKeyShare,
			msg: &p2pmsg.DecryptionKeyShare{
				InstanceID:  config.InstanceID,
				Eon:         eon,
				EpochID:     epochID.Bytes(),
				KeyperIndex: keyperIndex,
				Share:       keyshare,
			},
		},
		{
			name:      "invalid decryption key share wrong epoch",
			valid:     false,
			validator: validateDecryptionKeyShare,
			msg: &p2pmsg.DecryptionKeyShare{
				InstanceID:  config.InstanceID,
				Eon:         eon,
				EpochID:     wrongEpochID.Bytes(),
				KeyperIndex: keyperIndex,
				Share:       keyshare,
			},
		},
		{
			name:      "invalid decryption key share wrong instance ID",
			valid:     false,
			validator: validateDecryptionKeyShare,
			msg: &p2pmsg.DecryptionKeyShare{
				InstanceID:  config.InstanceID + 1,
				Eon:         eon,
				EpochID:     epochID.Bytes(),
				KeyperIndex: keyperIndex,
				Share:       keyshare,
			},
		},
		{
			name:      "invalid decryption key share wrong keyper index",
			valid:     false,
			validator: validateDecryptionKeyShare,
			msg: &p2pmsg.DecryptionKeyShare{
				InstanceID:  config.InstanceID,
				Eon:         eon,
				EpochID:     epochID.Bytes(),
				KeyperIndex: keyperIndex + 1,
				Share:       keyshare,
			},
		},
	}
	for _, tc := range tests {
		topic := tc.msg.Topic()
		t.Run(tc.name, func(t *testing.T) {
			pubsubMessage, err := makePubSubMessage(tc.msg, topic)
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
	_, dbpool, closedb := testdb.NewKeyperTestDB(ctx, t)
	defer closedb()

	p2pHandler := p2p.New(p2p.Config{})
	kpr := New(config.Address, config.InstanceID, dbpool)

	collatorKey1, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	collatorAddress1 := ethcrypto.PubkeyToAddress(collatorKey1.PublicKey)

	collatorKey2, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	collatorAddress2 := ethcrypto.PubkeyToAddress(collatorKey2.PublicKey)

	// Make a db with collator 1 from a certain block and collator 2 afterwards
	activationBlk1 := uint64(0)
	epochID1, _ := epochid.BigToEpochID(common.Big0)
	activationBlk2 := uint64(123)
	epochID2, _ := epochid.BigToEpochID(common.Big1)
	assert.NilError(t, err)
	collator1 := shdb.EncodeAddress(collatorAddress1)
	collator2 := shdb.EncodeAddress(collatorAddress2)
	err = chainobsdb.New(dbpool).InsertChainCollator(ctx, chainobsdb.InsertChainCollatorParams{
		ActivationBlockNumber: int64(activationBlk1),
		Collator:              collator1,
	})
	assert.NilError(t, err)
	err = chainobsdb.New(dbpool).InsertChainCollator(ctx, chainobsdb.InsertChainCollatorParams{
		ActivationBlockNumber: int64(activationBlk2),
		Collator:              collator2,
	})
	assert.NilError(t, err)

	var peerID peer.ID

	validateDecryptionTrigger := p2p.AddValidator(p2pHandler, kpr.validateDecryptionTrigger)

	tests := []struct {
		name        string
		valid       bool
		instanceID  uint64
		epochID     epochid.EpochID
		blockNumber uint64
		privKey     *ecdsa.PrivateKey
	}{
		{
			name:        "valid trigger collator 1",
			valid:       true,
			instanceID:  config.InstanceID,
			epochID:     epochID1,
			blockNumber: activationBlk1,
			privKey:     collatorKey1,
		},
		{
			name:        "valid trigger collator 2",
			valid:       true,
			instanceID:  config.InstanceID,
			epochID:     epochID2,
			blockNumber: activationBlk2,
			privKey:     collatorKey2,
		},
		{
			name:        "invalid trigger wrong collator 1",
			valid:       false,
			instanceID:  config.InstanceID,
			epochID:     epochID2,
			blockNumber: activationBlk2,
			privKey:     collatorKey1,
		},
		{
			name:        "invalid trigger wrong collator 2",
			valid:       false,
			instanceID:  config.InstanceID,
			epochID:     epochID1,
			blockNumber: activationBlk1,
			privKey:     collatorKey2,
		},
		{
			name:        "invalid trigger wrong instanceID",
			valid:       false,
			instanceID:  config.InstanceID + 1,
			epochID:     epochID1,
			blockNumber: activationBlk1,
			privKey:     collatorKey1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := p2pmsg.NewSignedDecryptionTrigger(
				tc.instanceID,
				tc.epochID,
				tc.blockNumber,
				[]byte{},
				tc.privKey,
			)
			assert.NilError(t, err)
			topic := msg.Topic()
			pubsubMsg, err := makePubSubMessage(msg, topic)
			if err != nil {
				t.Fatalf("Error in makePubSubMessage: %s", err)
			}
			assert.Equal(t, validateDecryptionTrigger(ctx, peerID, pubsubMsg), tc.valid,
				"validate failed valid=%t", tc.valid)
		})
	}
}

// makePubSubMessage makes a pubsub.Message corresponding to the type received by gossip validators.
func makePubSubMessage(message p2pmsg.Message, topic string) (*pubsub.Message, error) {
	messageBytes, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	pubsubMessage := pubsub.Message{
		Message: &pb.Message{
			Data:  messageBytes,
			Topic: &topic,
		},
		ReceivedFrom:  "",
		ValidatorData: nil,
	}

	return &pubsubMessage, nil
}
