package decryptor

import (
	"bytes"
	"context"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shuttermint/decryptor/dcrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func TestInvalidInputTypesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	_, err := handleInput(ctx, db, 5)
	if err == nil {
		t.Errorf("no error when receiving invalid type")
	}
}

func TestInsertDecryptionKeyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	m := &shmsg.DecryptionKey{
		EpochID: 100,
		Key:     []byte("hello"),
	}
	msgs, err := handleInput(ctx, db, m)
	assert.NilError(t, err)

	mStored, err := db.GetDecryptionKey(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	assert.NilError(t, err)
	assert.Check(t, medley.BytesEpochIDToUint64(mStored.EpochID) == m.EpochID)
	assert.Check(t, bytes.Equal(mStored.Key, m.Key))

	assert.Check(t, len(msgs) == 0)

	m2 := &shmsg.DecryptionKey{
		EpochID: 100,
		Key:     []byte("hello2"),
	}
	msgs, err = handleInput(ctx, db, m2)
	assert.NilError(t, err)

	m2Stored, err := db.GetDecryptionKey(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	assert.NilError(t, err)
	assert.Check(t, bytes.Equal(m2Stored.Key, m.Key))

	assert.Check(t, len(msgs) == 0)
}

func TestInsertCipherBatchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	m := &shmsg.CipherBatch{
		EpochID: 100,
		Data:    []byte("hello"),
	}
	msgs, err := handleInput(ctx, db, m)
	assert.NilError(t, err)

	mStored, err := db.GetCipherBatch(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	assert.NilError(t, err)
	assert.Check(t, medley.BytesEpochIDToUint64(mStored.EpochID) == m.EpochID)
	assert.Check(t, bytes.Equal(mStored.Data, m.Data))

	assert.Check(t, len(msgs) == 0)

	m2 := &shmsg.CipherBatch{
		EpochID: 100,
		Data:    []byte("hello2"),
	}
	msgs, err = handleInput(ctx, db, m2)
	assert.NilError(t, err)

	m2Stored, err := db.GetCipherBatch(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	assert.NilError(t, err)
	assert.Check(t, bytes.Equal(m2Stored.Data, m.Data))

	assert.Check(t, len(msgs) == 0)
}

func TestHandleEpochIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	cipherBatchMsg := &shmsg.CipherBatch{
		EpochID: 123,
		Data:    []byte("hello"),
	}
	msgs, err := handleInput(ctx, db, cipherBatchMsg)
	assert.NilError(t, err)
	assert.Check(t, len(msgs) == 0)

	keyMsg := &shmsg.DecryptionKey{
		EpochID: 123,
		Key:     []byte("hello"),
	}
	msgs, err = handleInput(ctx, db, keyMsg)
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
