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

	d := NewDecryptor(db)
	err := d.handleInput(ctx, 5)
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

	d := NewDecryptor(db)
	m := &shmsg.DecryptionKey{
		EpochID: 100,
		Key:     []byte("hello"),
	}
	err := d.handleInput(ctx, m)
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}

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
	err = d.handleInput(ctx, m2)
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}
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

	d := NewDecryptor(db)
	m := &shmsg.CipherBatch{
		EpochID: 100,
		Data:    []byte("hello"),
	}
	err := d.handleInput(ctx, m)
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}

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
	err = d.handleInput(ctx, m2)
	if err != nil {
		t.Fatalf("error handling input: %v", err)
	}
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

	d := Decryptor{
		db:            db,
		inputChannel:  make(<-chan interface{}),
		outputChannel: make(chan interface{}, 2),
	}

	cipherBatchMsg := &shmsg.CipherBatch{
		EpochID: 123,
		Data:    []byte("hello"),
	}
	if err := d.handleInput(ctx, cipherBatchMsg); err != nil {
		t.Fatalf("error handling input: %v", err)
	}

	keyMsg := &shmsg.DecryptionKey{
		EpochID: 123,
		Key:     []byte("hello"),
	}
	if err := d.handleInput(ctx, keyMsg); err != nil {
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

	select {
	case outputMessage := <-d.outputChannel:
		msg, ok := outputMessage.(*shmsg.AggregatedDecryptionSignature)
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
	default:
		t.Errorf("no message sent")
	}
}
