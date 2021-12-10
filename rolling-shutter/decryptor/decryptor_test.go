package decryptor

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"google.golang.org/protobuf/proto"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/medley/bitfield"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
	"github.com/shutter-network/shutter/shuttermint/p2p"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

// Add decryptor addresses to given db.
// Uses index of signingKeys map as decryptor index and arbitrary unique addresses.
func populateDBWithDecryptors(ctx context.Context, t *testing.T, db *dcrdb.Queries, signingKeys map[int32]*shbls.SecretKey) {
	t.Helper()
	for i, signingKey := range signingKeys {
		arbitraryAddress := fmt.Sprint(signingKey)
		err := db.InsertDecryptorSetMember(ctx, dcrdb.InsertDecryptorSetMemberParams{
			ActivationBlockNumber: 0,
			Index:                 i,
			Address:               arbitraryAddress,
		})
		assert.NilError(t, err)
		err = db.InsertDecryptorIdentity(ctx, dcrdb.InsertDecryptorIdentityParams{
			Address:        arbitraryAddress,
			BlsPublicKey:   shbls.SecretToPublicKey(signingKey).Marshal(),
			BlsSignature:   []byte{},
			SignatureValid: true,
		})
		assert.NilError(t, err)
	}
}

func TestCipherBatchValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	config := newTestConfig(t)
	d := New(config)
	d.db = db

	err := d.db.InsertChainCollator(ctx, dcrdb.InsertChainCollatorParams{
		ActivationBlockNumber: 1,
		Collator:              shdb.EncodeAddress(config.EthereumAddress()),
	})
	assert.NilError(t, err)
	err = d.db.InsertChainCollator(ctx, dcrdb.InsertChainCollatorParams{
		ActivationBlockNumber: 20,
		Collator:              "0x0000000000000000000000000000000000000000",
	})
	assert.NilError(t, err)

	txs := [][]byte{
		[]byte("tx1"),
		[]byte("tx2"),
		[]byte("tx3"),
	}
	epochID := epochid.New(0, 1)

	trigger, err := shmsg.NewSignedDecryptionTrigger(d.Config.InstanceID, epochID, txs, config.EthereumKey)
	assert.NilError(t, err)
	triggerBadInstanceID, err := shmsg.NewSignedDecryptionTrigger(
		d.Config.InstanceID+2,
		epochID,
		txs,
		config.EthereumKey,
	)
	assert.NilError(t, err)
	wrongCollatorEpochID := epochid.New(0, 20)
	triggerWrongCollator, err := shmsg.NewSignedDecryptionTrigger(
		d.Config.InstanceID,
		wrongCollatorEpochID,
		txs,
		config.EthereumKey,
	)
	assert.NilError(t, err)
	triggerBadHash, err := shmsg.NewSignedDecryptionTrigger(
		d.Config.InstanceID,
		wrongCollatorEpochID,
		append(txs, []byte("tx4")),
		config.EthereumKey,
	)
	assert.NilError(t, err)
	triggerBadSignature := &shmsg.DecryptionTrigger{
		InstanceID:       trigger.InstanceID,
		EpochID:          trigger.EpochID,
		TransactionsHash: trigger.TransactionsHash,
		Signature:        []byte("badsignature"),
	}

	var peerID peer.ID
	tests := []struct {
		name    string
		valid   bool
		trigger *shmsg.DecryptionTrigger
	}{
		{
			name:    "valid cipher batch",
			valid:   true,
			trigger: trigger,
		},
		{
			name:    "invalid cipher batch instance ID",
			valid:   false,
			trigger: triggerBadInstanceID,
		},
		{
			name:    "invalid cipher batch collator",
			valid:   false,
			trigger: triggerWrongCollator,
		},
		{
			name:    "invalid cipher batch transaction hash",
			valid:   false,
			trigger: triggerBadHash,
		},
		{
			name:    "invalid cipher batch signature",
			valid:   false,
			trigger: triggerBadSignature,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cipherBatch := &shmsg.CipherBatch{
				DecryptionTrigger: tc.trigger,
				Transactions:      txs,
			}
			pubsubMessage, err := makePubSubMessage(cipherBatch, cipherBatch.Topic())
			if err != nil {
				t.Fatalf("Error in makePubSubMessage: %s", err)
			}
			assert.Equal(t, d.validateCipherBatch(ctx, peerID, pubsubMessage), tc.valid,
				"validate failed valid=%t msg=%+v type=%T", tc.valid, cipherBatch, cipherBatch)
		})
	}
}

func TestSignatureValidatorsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	var peerID peer.ID
	config := newTestConfig(t)
	d := New(config)
	d.db = db

	signingKey2, _, err := shbls.RandomKeyPair(rand.Reader)
	assert.NilError(t, err)
	populateDBWithDecryptors(ctx, t, db, map[int32]*shbls.SecretKey{0: config.SigningKey, 1: signingKey2})

	validators := d.makeMessagesValidators()

	validHash := common.BytesToHash([]byte("Hello"))
	wrongHash := common.BytesToHash([]byte("Not Hello"))
	validSignature := shbls.Sign(validHash.Bytes(), config.SigningKey)

	validSignature2 := shbls.Sign(validHash.Bytes(), signingKey2)
	aggregatedSignature := shbls.AggregateSignatures([]*shbls.Signature{validSignature, validSignature2})

	tests := []struct {
		name  string
		valid bool
		msg   shmsg.P2PMessage
	}{
		{
			name:  "valid signature",
			valid: true,
			msg: &shmsg.DecryptionSignature{
				InstanceID:     d.Config.InstanceID,
				Signature:      validSignature.Marshal(),
				SignedHash:     validHash.Bytes(),
				SignerBitfield: bitfield.MakeBitfieldFromIndex(0),
			},
		},
		{
			name:  "invalid signature two signers",
			valid: false,
			msg: &shmsg.DecryptionSignature{
				InstanceID:     d.Config.InstanceID,
				Signature:      aggregatedSignature.Marshal(),
				SignedHash:     validHash.Bytes(),
				SignerBitfield: bitfield.MakeBitfieldFromIndex(0, 1),
			},
		},
		{
			name:  "invalid signature instance id",
			valid: false,
			msg: &shmsg.DecryptionSignature{
				InstanceID: d.Config.InstanceID - 1,
			},
		},
		{
			name:  "invalid signature hash",
			valid: false,
			msg: &shmsg.DecryptionSignature{
				InstanceID: d.Config.InstanceID,
				Signature:  validSignature.Marshal(),
				SignedHash: wrongHash.Bytes(),
			},
		},
		{
			name:  "valid aggregated signature one signer",
			valid: true,
			msg: &shmsg.AggregatedDecryptionSignature{
				InstanceID:          d.Config.InstanceID,
				AggregatedSignature: validSignature.Marshal(),
				SignedHash:          validHash.Bytes(),
				SignerBitfield:      bitfield.MakeBitfieldFromIndex(0),
			},
		},
		{
			name:  "valid aggregated signature two signers",
			valid: true,
			msg: &shmsg.AggregatedDecryptionSignature{
				InstanceID:          d.Config.InstanceID,
				AggregatedSignature: aggregatedSignature.Marshal(),
				SignedHash:          validHash.Bytes(),
				SignerBitfield:      bitfield.MakeBitfieldFromIndex(0, 1),
			},
		},
		{
			name:  "invalid aggregated signature instance id",
			valid: false,
			msg: &shmsg.AggregatedDecryptionSignature{
				InstanceID: d.Config.InstanceID - 1,
			},
		},
		{
			name:  "invalid aggregated signature hash",
			valid: false,
			msg: &shmsg.AggregatedDecryptionSignature{
				InstanceID:          d.Config.InstanceID,
				AggregatedSignature: validSignature.Marshal(),
				SignedHash:          wrongHash.Bytes(),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pubsubMessage, err := makePubSubMessage(tc.msg, tc.msg.Topic())
			if err != nil {
				t.Fatalf("Error in makePubSubMessage: %s", err)
			}
			validate := validators[pubsubMessage.GetTopic()]
			assert.Assert(t, validate != nil)
			assert.Equal(t, validate(ctx, peerID, pubsubMessage), tc.valid,
				"validate failed valid=%t msg=%+v type=%T", tc.valid, tc.msg, tc.msg)
		})
	}
}

func TestDecryptionKeyValidatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	var peerID peer.ID
	config := newTestConfig(t)
	d := New(config)
	d.db = db

	tkg := medley.NewTestKeyGenerator(t, 1, 1)

	err := db.InsertEonPublicKey(ctx, dcrdb.InsertEonPublicKeyParams{
		ActivationBlockNumber: 0,
		EonPublicKey:          tkg.EonPublicKey(0).Marshal(),
	})
	assert.NilError(t, err)

	validSecretKey := tkg.EpochSecretKey(0).Marshal()
	invalidSecretKey := tkg.EpochSecretKey(1).Marshal()

	tests := []struct {
		name  string
		valid bool
		msg   shmsg.P2PMessage
	}{
		{
			name:  "valid decryption key",
			valid: true,
			msg: &shmsg.DecryptionKey{
				InstanceID: d.Config.InstanceID,
				Key:        validSecretKey,
				EpochID:    0,
			},
		},
		{
			name:  "invalid decryption key wrong epoch",
			valid: false,
			msg: &shmsg.DecryptionKey{
				InstanceID: d.Config.InstanceID,
				Key:        invalidSecretKey,
				EpochID:    0,
			},
		},
		{
			name:  "invalid decryption key instance ID",
			valid: false,
			msg: &shmsg.DecryptionKey{
				InstanceID: d.Config.InstanceID + 1,
				Key:        validSecretKey,
				EpochID:    0,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pubsubMessage, err := makePubSubMessage(tc.msg, tc.msg.Topic())
			if err != nil {
				t.Fatalf("Error in makePubSubMessage: %s", err)
			}
			assert.Equal(t, d.validateDecryptionKey(ctx, peerID, pubsubMessage), tc.valid,
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
