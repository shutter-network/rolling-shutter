package decryptor

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func newTestConfig(t *testing.T) Config {
	t.Helper()

	p2pKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	assert.NilError(t, err)
	signingKey, _, err := shbls.RandomKeyPair(rand.Reader)
	assert.NilError(t, err)
	return Config{
		ListenAddress:  nil,
		PeerMultiaddrs: nil,

		DatabaseURL: "",

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
		StartEpochID: medley.Uint64EpochIDToBytes(0),
		EonPublicKey: tkg.EonPublicKey(0).Marshal(),
	})
	assert.NilError(t, err)

	// send an epoch secret key and check that it's stored in the db
	m := &decryptionKey{
		epochID: 0,
		key:     tkg.EpochSecretKey(0),
	}
	msgs, err := handleDecryptionKeyInput(ctx, config, db, m)
	assert.NilError(t, err)

	mStored, err := db.GetDecryptionKey(ctx, medley.Uint64EpochIDToBytes(m.epochID))
	assert.NilError(t, err)
	assert.Check(t, medley.BytesEpochIDToUint64(mStored.EpochID) == m.epochID)
	keyBytes, _ := m.key.GobEncode()
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
		EpochID:      100,
		Transactions: [][]byte{[]byte("tx1"), []byte("tx2")},
	}
	msgs, err := handleCipherBatchInput(ctx, config, db, m)
	assert.NilError(t, err)

	mStored, err := db.GetCipherBatch(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	assert.NilError(t, err)
	assert.Check(t, medley.BytesEpochIDToUint64(mStored.EpochID) == m.EpochID)
	assert.DeepEqual(t, mStored.Transactions, m.Transactions)
	assert.Check(t, len(msgs) == 0)

	m2 := &cipherBatch{
		EpochID:      100,
		Transactions: [][]byte{[]byte("tx3")},
	}
	msgs, err = handleCipherBatchInput(ctx, config, db, m2)
	assert.NilError(t, err)

	m2Stored, err := db.GetCipherBatch(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
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

	bitfield := makeBitfieldFromIndex(1)
	bitfield2 := makeBitfieldFromIndex(2)
	hash := common.BytesToHash([]byte("Hello"))
	signature := &decryptionSignature{
		epochID:        0,
		instanceID:     config.InstanceID,
		signedHash:     hash,
		signature:      shbls.Sign(hash.Bytes(), config.SigningKey),
		SignerBitfield: bitfield,
	}
	signature2 := &decryptionSignature{
		epochID:        0,
		instanceID:     config.InstanceID,
		signedHash:     hash,
		signature:      shbls.Sign(hash.Bytes(), signingKey2),
		SignerBitfield: bitfield2,
	}

	tests := []struct {
		name    string
		config  Config
		inputs  []*decryptionSignature
		outputs []*shmsg.AggregatedDecryptionSignature
	}{
		{
			name:    "single signature required",
			config:  config,
			inputs:  []*decryptionSignature{signature},
			outputs: []*shmsg.AggregatedDecryptionSignature{{InstanceID: config.InstanceID, SignedHash: hash.Bytes(), SignerBitfield: bitfield}},
		},
		{
			name:    "two signatures required",
			config:  configTwoRequiredSignatures,
			inputs:  []*decryptionSignature{signature},
			outputs: []*shmsg.AggregatedDecryptionSignature{nil},
		},
		{
			name:   "two signatures required two provided",
			config: configTwoRequiredSignatures,
			inputs: []*decryptionSignature{signature, signature2},
			outputs: []*shmsg.AggregatedDecryptionSignature{nil, {
				InstanceID: configTwoRequiredSignatures.InstanceID, SignedHash: hash.Bytes(),
				SignerBitfield: makeBitfieldFromArray([]int32{config.SignerIndex, 2}),
			}},
		},
	}

	for _, test := range tests {
		db, closedb := medley.NewDecryptorTestDB(ctx, t)
		populateDBWithDecryptors(ctx, t, db, map[int32]*shbls.SecretKey{config.SignerIndex: config.SigningKey, 2: signingKey2})
		t.Run(test.name, func(t *testing.T) {
			for i, input := range test.inputs {
				msgs, err := handleSignatureInput(ctx, test.config, db, input)
				assert.NilError(t, err)
				isOutputNill := test.outputs[i] == nil
				if isOutputNill {
					assert.Check(t, len(msgs) == 0)
				} else {
					assert.Check(t, len(msgs) == 1)
					msg, ok := msgs[0].(*shmsg.AggregatedDecryptionSignature)
					assert.Check(t, ok, "wrong message type")
					assert.Equal(t, msg.InstanceID, test.outputs[i].InstanceID)
					assert.Check(t, bytes.Equal(msg.SignedHash, test.outputs[i].SignedHash))
					assert.Check(t, bytes.Equal(msg.SignerBitfield, test.outputs[i].SignerBitfield))
				}
			}
		})
		closedb()
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

	bitfield := makeBitfieldFromIndex(1)
	bitfield2 := makeBitfieldFromIndex(2)
	hash := common.BytesToHash([]byte("Hello"))
	signature := &decryptionSignature{
		epochID:        0,
		instanceID:     config.InstanceID,
		signedHash:     hash,
		signature:      shbls.Sign(hash.Bytes(), config.SigningKey),
		SignerBitfield: bitfield,
	}
	signature2 := &decryptionSignature{
		epochID:        0,
		instanceID:     config.InstanceID,
		signedHash:     hash,
		signature:      shbls.Sign(hash.Bytes(), signingKey2),
		SignerBitfield: bitfield2,
	}

	msgs, err := handleSignatureInput(ctx, config, db, signature)
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 1)

	signatureStored, err := db.GetAggregatedSignature(ctx, signature.signedHash.Bytes())
	assert.NilError(t, err)
	assert.Equal(t, signature.epochID, medley.BytesEpochIDToUint64(signatureStored.EpochID))
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
		StartEpochID: medley.Uint64EpochIDToBytes(0),
		EonPublicKey: tkg.EonPublicKey(0).Marshal(),
	})
	assert.NilError(t, err)

	cipherBatchMsg := &cipherBatch{
		EpochID:      0,
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
		EpochID:         medley.Uint64EpochIDToBytes(cipherBatchMsg.EpochID),
		SignersBitfield: makeBitfieldFromIndex(config.SignerIndex),
	})
	assert.NilError(t, err)

	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*shmsg.DecryptionSignature)
	assert.Check(t, ok, "wrong message type")
	assert.Equal(
		t,
		medley.BytesEpochIDToUint64(storedDecryptionKey.EpochID),
		msg.EpochID,
	)
	assert.Check(t, bytes.Equal(storedDecryptionKey.SignedHash, msg.SignedHash))
	assert.Check(t, bytes.Equal(storedDecryptionKey.Signature, msg.Signature))
}
