package decryptor

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"

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

		P2PKey:     p2pKey,
		SigningKey: signingKey,
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
	tkg := medley.NewTestKeyGenerator(t)

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

	// send a wrong epoch secret key (e.g., one for a wrong epoch) and check that there's an error
	m2 := &decryptionKey{
		epochID: 1,
		key:     tkg.EpochSecretKey(2),
	}
	_, err = handleDecryptionKeyInput(ctx, config, db, m2)
	assert.Check(t, err != nil)
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

func TestHandleEpochIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)
	tkg := medley.NewTestKeyGenerator(t)

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

	// TODO: handle signer index
	storedDecryptionKey,
		err := db.GetDecryptionSignature(ctx, dcrdb.GetDecryptionSignatureParams{
		EpochID:     medley.Uint64EpochIDToBytes(cipherBatchMsg.EpochID),
		SignerIndex: 0,
	})
	assert.NilError(t, err)

	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*shmsg.AggregatedDecryptionSignature)
	assert.Check(t, ok, "wrong message type")
	assert.Equal(
		t,
		medley.BytesEpochIDToUint64(storedDecryptionKey.EpochID),
		msg.EpochID,
	)
	assert.Check(t, bytes.Equal(storedDecryptionKey.SignedHash, msg.SignedHash))
	assert.Check(t, bytes.Equal(storedDecryptionKey.Signature, msg.AggregatedSignature))
}
