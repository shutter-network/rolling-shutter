package decryptor

import (
	"bytes"
	"context"
	"testing"

	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func TestInvalidInputTypes(t *testing.T) {
	ctx := context.Background()
	db, closedb := medley.NewDecryptorTestDB(ctx, t)
	defer closedb()

	d := NewDecryptor(db)
	err := d.handleInput(ctx, 5)
	if err == nil {
		t.Errorf("no error when receiving invalid type")
	}
}

func TestInsertDecryptionKey(t *testing.T) {
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

	mStored, err := db.GetDecryptionKey(ctx, int64(m.EpochID))
	if err != nil {
		t.Fatalf("error retrieving decryption key: %v", err)
	}

	if mStored.EpochID != int64(m.EpochID) {
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
	m2Stored, err := db.GetDecryptionKey(ctx, int64(m.EpochID))
	if err != nil {
		t.Fatalf("error retrieving decryption key: %v", err)
	}
	if !bytes.Equal(m2Stored.Key, m.Key) {
		t.Errorf("inserting another decryption key changed existing one")
	}
}

func TestInsertCipherBatch(t *testing.T) {
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

	mStored, err := db.GetCipherBatch(ctx, int64(m.EpochID))
	if err != nil {
		t.Fatalf("error retrieving cipher batch: %v", err)
	}

	if mStored.EpochID != int64(m.EpochID) {
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
	m2Stored, err := db.GetCipherBatch(ctx, int64(m.EpochID))
	if err != nil {
		t.Fatalf("error retrieving data: %v", err)
	}
	if !bytes.Equal(m2Stored.Data, m.Data) {
		t.Errorf("inserting data twice changed existing one")
	}
}
