package collator

import (
	"context"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shuttermint/collator/cltrdb"
	"github.com/shutter-network/shutter/shuttermint/medley"
	"github.com/shutter-network/shutter/shuttermint/medley/epochid"
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

	nextEpochID := epochid.New(41, 101)
	err := db.SetNextEpochID(ctx, shdb.EncodeUint64(nextEpochID))
	assert.NilError(t, err)

	encryptedTX := []byte("foobar")
	err = db.InsertTx(ctx, cltrdb.InsertTxParams{
		TxID:        []byte{'a'},
		EpochID:     shdb.EncodeUint64(nextEpochID),
		EncryptedTx: encryptedTX,
	})
	assert.NilError(t, err)

	transactionsHash := shmsg.HashTransactions([][]byte{encryptedTX})

	msgs, err := startNextEpoch(ctx, config, db, 102)
	assert.NilError(t, err)

	// make sure decryption trigger is stored in db
	stored, err := db.GetTrigger(ctx, shdb.EncodeUint64(nextEpochID))
	assert.NilError(t, err)
	assert.Equal(t, shdb.DecodeUint64(stored.EpochID), nextEpochID)
	assert.DeepEqual(t, stored.BatchHash, transactionsHash)

	batchMsg := msgs[0].(*shmsg.CipherBatch)
	assert.Equal(t, batchMsg.DecryptionTrigger.InstanceID, config.InstanceID)
	assert.Equal(t, batchMsg.DecryptionTrigger.EpochID, nextEpochID)
	assert.DeepEqual(t, batchMsg.Transactions, [][]byte{{'f', 'o', 'o', 'b', 'a', 'r'}})

	// make sure output is trigger message
	triggerMsg := msgs[1].(*shmsg.DecryptionTrigger)
	assert.Equal(t, triggerMsg.InstanceID, config.InstanceID)
	assert.Equal(t, triggerMsg.EpochID, nextEpochID)
	assert.DeepEqual(t, triggerMsg.TransactionsHash, transactionsHash)
	address := ethcrypto.PubkeyToAddress(config.EthereumKey.PublicKey)
	signatureCorrect, err := triggerMsg.VerifySignature(address)
	assert.NilError(t, err)
	assert.Check(t, signatureCorrect)
}
