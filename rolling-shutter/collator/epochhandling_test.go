package collator

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/shdb"
	"github.com/shutter-network/shutter/shuttermint/shmsg"
)

func newTestConfig(t *testing.T) Config {
	t.Helper()

	ethereumKey, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	return Config{
		EthereumKey: ethereumKey,
		InstanceID:  123,
	}
}

func TestDecryptionTriggerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewCollatorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)

	nextEpochID := uint64(0)
	batch := &shmsg.CipherBatch{
		InstanceID:   config.InstanceID,
		EpochID:      nextEpochID,
		Transactions: [][]byte{},
	}
	hash := sha3.Sum256(bytes.Join(batch.Transactions, []byte{}))
	msg, err := makeDecryptionTrigger(ctx, config, db, batch)
	assert.NilError(t, err)

	// make sure decryption trigger is stored in db
	stored, err := db.GetTrigger(ctx, shdb.EncodeUint64(nextEpochID))
	assert.NilError(t, err)
	assert.Equal(t, shdb.DecodeUint64(stored.EpochID), nextEpochID)
	assert.Check(t, bytes.Equal(stored.BatchHash, hash[:]))

	// make sure output is trigger message
	assert.Equal(t, msg.InstanceID, config.InstanceID)
	assert.Equal(t, msg.EpochID, nextEpochID)
	assert.Check(t, bytes.Equal(msg.TransactionsHash, hash[:]))
	address := ethcrypto.PubkeyToAddress(config.EthereumKey.PublicKey)
	signatureCorrect, err := msg.VerifySignature(address)
	assert.NilError(t, err)
	assert.Check(t, signatureCorrect)
}

func TestBatchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewCollatorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)

	msg, err := makeBatch(ctx, config, db)
	assert.NilError(t, err)

	// make sure cipher batch is stored in db
	stored, err := db.GetBatch(ctx, shdb.EncodeUint64(0))
	assert.NilError(t, err)
	assert.Equal(t, len(stored.Transactions), 0)

	// make sure output is cipher batch
	assert.Equal(t, msg.InstanceID, config.InstanceID)
	assert.Equal(t, msg.EpochID, uint64(0))
	assert.Equal(t, len(msg.Transactions), 0)
}

func TestHandleEpochIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, closedb := medley.NewCollatorTestDB(ctx, t)
	defer closedb()
	config := newTestConfig(t)

	msgs, err := handleEpoch(ctx, config, db)
	assert.NilError(t, err)

	assert.Equal(t, len(msgs), 2)
	assert.Check(t, reflect.TypeOf(msgs[0]) != reflect.TypeOf(msgs[1]))

	for _, msg := range msgs {
		switch typedMsg := msg.(type) {
		case *shmsg.CipherBatch:
			assert.Equal(t, typedMsg.InstanceID, config.InstanceID)
			assert.Equal(t, typedMsg.EpochID, uint64(0))
		case *shmsg.DecryptionTrigger:
			assert.Equal(t, typedMsg.InstanceID, config.InstanceID)
			assert.Equal(t, typedMsg.EpochID, uint64(0))
		default:
			t.Errorf("invalid message type %T", msg)
		}
	}
}
