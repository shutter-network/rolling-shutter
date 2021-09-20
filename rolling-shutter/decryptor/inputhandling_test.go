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
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}

	assert.Check(t, len(msgs) == 0)
	mStored, err := db.GetDecryptionKey(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	if err != nil {
		t.Fatalf("error retrieving decryption key: %v", err)
	}

	if medley.BytesEpochIDToUint64(mStored.EpochID) != m.EpochID {
		t.Errorf("wrong epoch id")
	}
	if !bytes.Equal(mStored.Key, m.Key) {
		t.Errorf("wrong key")
	}

	m2 := &shmsg.DecryptionKey{
		EpochID: 100,
		Key:     []byte("hello2"),
	}
	msgs, err = handleInput(ctx, db, m2)
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}
	assert.Check(t, len(msgs) == 0)
	m2Stored, err := db.GetDecryptionKey(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	if err != nil {
		t.Fatalf("error retrieving decryption key: %v", err)
	}
	if !bytes.Equal(m2Stored.Key, m.Key) {
		t.Errorf("inserting another decryption key changed existing one")
	}
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
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}

	assert.Check(t, len(msgs) == 0)
	mStored, err := db.GetCipherBatch(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	if err != nil {
		t.Fatalf("error retrieving cipher batch: %v", err)
	}

	if medley.BytesEpochIDToUint64(mStored.EpochID) != m.EpochID {
		t.Errorf("wrong epoch id")
	}
	if !bytes.Equal(mStored.Data, m.Data) {
		t.Errorf("wrong data")
	}

	m2 := &shmsg.CipherBatch{
		EpochID: 100,
		Data:    []byte("hello2"),
	}
	msgs, err = handleInput(ctx, db, m2)
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}
	assert.Check(t, len(msgs) == 0)
	m2Stored, err := db.GetCipherBatch(ctx, medley.Uint64EpochIDToBytes(m.EpochID))
	if err != nil {
		t.Fatalf("error retrieving data: %v", err)
	}
	if !bytes.Equal(m2Stored.Data, m.Data) {
		t.Errorf("inserting data twice changed existing one")
	}
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
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}
	assert.Check(t, len(msgs) == 0)

	assert.Check(t, len(msgs) == 0)
	keyMsg := &shmsg.DecryptionKey{
		EpochID: 123,
		Key:     []byte("hello"),
	}
	msgs, err = handleInput(ctx, db, keyMsg)
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}

	// TODO: handle signer index
	storedDecryptionKey,
		err := db.GetDecryptionSignature(ctx, dcrdb.GetDecryptionSignatureParams{
		EpochID:     medley.Uint64EpochIDToBytes(cipherBatchMsg.EpochID),
		SignerIndex: 0,
	})
	if err != nil {
		t.Fatalf("error retrieving cipher batch: %v", err)
	}

	assert.Check(t, len(msgs) == 1)
	msg, ok := msgs[0].(*shmsg.AggregatedDecryptionSignature)
	if !ok {
		t.Errorf("wrong type")
	}
	assert.Equal(
		t,
		medley.BytesEpochIDToUint64(storedDecryptionKey.EpochID),
		msg.EpochID,
		"stored and output epoch id do not match",
	)
	if !bytes.Equal(storedDecryptionKey.SignedHash, msg.SignedHash) {
		t.Errorf("stored and output signed hash do not match")
	}
	if !bytes.Equal(storedDecryptionKey.Signature, msg.AggregatedSignature) {
		t.Errorf("stored and output aggregated signature do not match")
	}
}
