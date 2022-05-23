package collator

import (
	"context"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func TestDecryptionTriggerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	db, _, closedb := testdb.NewCollatorTestDB(ctx, t)
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

	// make sure output is trigger message
	assert.Equal(t, len(msgs), 1)
	triggerMsg := msgs[0].(*shmsg.DecryptionTrigger)
	assert.Equal(t, triggerMsg.InstanceID, config.InstanceID)
	assert.Equal(t, triggerMsg.EpochID, nextEpochID)
	assert.DeepEqual(t, triggerMsg.TransactionsHash, transactionsHash)
	address := ethcrypto.PubkeyToAddress(config.EthereumKey.PublicKey)
	signatureCorrect, err := shmsg.VerifySignature(triggerMsg, address)
	assert.NilError(t, err)
	assert.Check(t, signatureCorrect)
}
