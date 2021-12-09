package decryptor

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	p2pcrypto "github.com/libp2p/go-libp2p-crypto"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/medley/bitfield"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func newTestConfig(t *testing.T) Config {
	t.Helper()

	ethereumKey, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	p2pKey, _, err := p2pcrypto.GenerateEd25519Key(rand.Reader)
	assert.NilError(t, err)
	signingKey, _, err := shbls.RandomKeyPair(rand.Reader)
	assert.NilError(t, err)
	return Config{
		ListenAddress:  nil,
		PeerMultiaddrs: nil,

		DatabaseURL: "",

		EthereumKey: ethereumKey,
		P2PKey:      p2pKey,
		SigningKey:  signingKey,
		SignerIndex: 1,

		InstanceID: 123,
	}
}

func TestInsertDecryptionKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)
	tkg := medley.NewTestKeyGenerator(t, 1, 1)

	err := db.InsertEonPublicKey(ctx, dcrdb.InsertEonPublicKeyParams{
		ActivationBlockNumber: 0,
		EonPublicKey:          tkg.EonPublicKey(0).Marshal(),
	})
	assert.NilError(t, err)

	// send an epoch secret key and check that it's stored in the db
	m := &decryptionKey{
		epochID: 0,
		key:     tkg.EpochSecretKey(0),
	}
	msgs, err := handleDecryptionKeyInput(ctx, config, db, m)
	assert.NilError(t, err)

	mStored, err := db.GetDecryptionKey(ctx, shdb.EncodeUint64(m.epochID))
	assert.NilError(t, err)
	assert.Check(t, shdb.DecodeUint64(mStored.EpochID) == m.epochID)
	keyBytes := m.key.Marshal()
	assert.Check(t, bytes.Equal(mStored.Key, keyBytes))

	assert.Check(t, len(msgs) == 0)
}

func TestInsertCipherBatchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)

	m := &cipherBatch{
		DecryptionTrigger: &shmsg.DecryptionTrigger{
			EpochID: 100,
		},
		Transactions: [][]byte{[]byte("tx1"), []byte("tx2")},
	}
	msgs, err := handleCipherBatchInput(ctx, config, db, m)
	assert.NilError(t, err)

	mStored, err := db.GetCipherBatch(ctx, shdb.EncodeUint64(m.DecryptionTrigger.EpochID))
	assert.NilError(t, err)
	assert.Check(t, shdb.DecodeUint64(mStored.EpochID) == m.DecryptionTrigger.EpochID)
	assert.DeepEqual(t, mStored.Transactions, m.Transactions)
	assert.Check(t, len(msgs) == 0)

	m2 := &cipherBatch{
		DecryptionTrigger: &shmsg.DecryptionTrigger{
			EpochID: 100,
		},
		Transactions: [][]byte{[]byte("tx3")},
	}
	msgs, err = handleCipherBatchInput(ctx, config, db, m2)
	assert.NilError(t, err)

	m2Stored, err := db.GetCipherBatch(ctx, shdb.EncodeUint64(m.DecryptionTrigger.EpochID))
	assert.NilError(t, err)
	assert.DeepEqual(t, m2Stored.Transactions, m.Transactions)

	assert.Check(t, len(msgs) == 0)
}

func TestHandleSignatureIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	config := newTestConfig(t)
	configTwoRequiredSignatures := config
	configTwoRequiredSignatures.RequiredSignatures = 2

	signingKey2, _, err := shbls.RandomKeyPair(rand.Reader)
	assert.NilError(t, err)

	bf := bitfield.MakeBitfieldFromIndex(1)
	bf2 := bitfield.MakeBitfieldFromIndex(2)
	hash := common.BytesToHash([]byte("Hello"))
	signature := &decryptionSignature{
		epochID:    0,
		instanceID: config.InstanceID,
		signedHash: hash,
		signature:  shbls.Sign(hash.Bytes(), config.SigningKey),
		signers:    bf,
	}
	signature2 := &decryptionSignature{
		epochID:    0,
		instanceID: config.InstanceID,
		signedHash: hash,
		signature:  shbls.Sign(hash.Bytes(), signingKey2),
		signers:    bf2,
	}

	tests := []struct {
		name    string
		config  Config
		inputs  []*decryptionSignature
		outputs []*shmsg.AggregatedDecryptionSignature
	}{
		{
			name:   "single signature required",
			config: config,
			inputs: []*decryptionSignature{
				signature,
			},
			outputs: []*shmsg.AggregatedDecryptionSignature{
				{
					InstanceID:     config.InstanceID,
					SignedHash:     hash.Bytes(),
					SignerBitfield: bf,
				},
			},
		},
		{
			name:   "two signatures required two provided",
			config: configTwoRequiredSignatures,
			inputs: []*decryptionSignature{
				signature,
				signature2,
			},
			outputs: []*shmsg.AggregatedDecryptionSignature{
				nil,
				{
					InstanceID:     configTwoRequiredSignatures.InstanceID,
					SignedHash:     hash.Bytes(),
					SignerBitfield: bitfield.MakeBitfieldFromIndex(config.SignerIndex, 2),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, closedb := medley.NewDecryptorTestDB(ctx, t)
			defer closedb()
			populateDBWithDecryptors(ctx, t, db, map[int32]*shbls.SecretKey{config.SignerIndex: config.SigningKey, 2: signingKey2})
			for i, input := range test.inputs {
				msgs, err := handleSignatureInput(ctx, test.config, db, input)
				assert.NilError(t, err)
				output := test.outputs[i]
				if output == nil {
					assert.Check(t, len(msgs) == 0)
				} else {
					assert.Check(t, len(msgs) == 1)
					msg, ok := msgs[0].(*shmsg.AggregatedDecryptionSignature)
					assert.Check(t, ok, "wrong message type")
					assert.Equal(t, msg.InstanceID, output.InstanceID)
					assert.Check(t, bytes.Equal(msg.SignedHash, output.SignedHash))
					assert.Check(t, bytes.Equal(msg.SignerBitfield, output.SignerBitfield))
				}
			}
		})
	}
}

func TestInsertAggregatedSignatureIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	config := newTestConfig(t)

	signingKey2, _, err := shbls.RandomKeyPair(rand.Reader)
	assert.NilError(t, err)

	populateDBWithDecryptors(ctx, t, db, map[int32]*shbls.SecretKey{config.SignerIndex: config.SigningKey, 2: signingKey2})

	bf := bitfield.MakeBitfieldFromIndex(1)
	bf2 := bitfield.MakeBitfieldFromIndex(2)
	hash := common.BytesToHash([]byte("Hello"))
	signature := &decryptionSignature{
		epochID:    0,
		instanceID: config.InstanceID,
		signedHash: hash,
		signature:  shbls.Sign(hash.Bytes(), config.SigningKey),
		signers:    bf,
	}
	signature2 := &decryptionSignature{
		epochID:    0,
		instanceID: config.InstanceID,
		signedHash: hash,
		signature:  shbls.Sign(hash.Bytes(), signingKey2),
		signers:    bf2,
	}

	msgs, err := handleSignatureInput(ctx, config, db, signature)
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 1)

	signatureStored, err := db.GetAggregatedSignature(ctx, signature.signedHash.Bytes())
	assert.NilError(t, err)
	assert.Equal(t, signature.epochID, shdb.DecodeUint64(signatureStored.EpochID))
	assert.Check(t, bytes.Equal(signature.signedHash.Bytes(), signatureStored.SignedHash))

	msgs, err = handleSignatureInput(ctx, config, db, signature2)
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)
}

func TestHandleEpochIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)
	config.RequiredSignatures = 2 // prevent generation of polluting signatures
	tkg := medley.NewTestKeyGenerator(t, 1, 1)

	err := db.InsertEonPublicKey(ctx, dcrdb.InsertEonPublicKeyParams{
		ActivationBlockNumber: 0,
		EonPublicKey:          tkg.EonPublicKey(0).Marshal(),
	})
	assert.NilError(t, err)

	cipherBatchMsg := &cipherBatch{
		DecryptionTrigger: &shmsg.DecryptionTrigger{
			EpochID: 0,
		},
		Transactions: [][]byte{[]byte("tx1")},
	}
	msgs, err := handleCipherBatchInput(ctx, config, db, cipherBatchMsg)
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)

	keyMsg := &decryptionKey{
		epochID: 0,
		key:     tkg.EpochSecretKey(0),
	}
	msgs, err = handleDecryptionKeyInput(ctx, config, db, keyMsg)
	assert.NilError(t, err)

	storedDecryptionKey,
		err := db.GetDecryptionSignature(ctx, dcrdb.GetDecryptionSignatureParams{
		EpochID:         shdb.EncodeUint64(cipherBatchMsg.DecryptionTrigger.EpochID),
		SignersBitfield: bitfield.MakeBitfieldFromIndex(config.SignerIndex),
	})
	assert.NilError(t, err)

	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*shmsg.DecryptionSignature)
	assert.Check(t, ok, "wrong message type")
	assert.Equal(
		t,
		shdb.DecodeUint64(storedDecryptionKey.EpochID),
		msg.EpochID,
	)
	assert.Check(t, bytes.Equal(storedDecryptionKey.SignedHash, msg.SignedHash))
	assert.Check(t, bytes.Equal(storedDecryptionKey.Signature, msg.Signature))
}
