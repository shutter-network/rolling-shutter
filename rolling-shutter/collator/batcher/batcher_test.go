package batcher

import (
	"context"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func DefaultTestParams() TestParams {
	return TestParams{
		GasLimit:       uint64(210000),
		BaseFee:        big.NewInt(1),
		InitialBalance: big.NewInt(1000000),
		TxGasTipCap:    big.NewInt(1),
		TxGasFeeCap:    big.NewInt(2),
		InitialEpochID: epochid.Uint64ToEpochID(2000),
	}
}

func TestRejectBadTransactionsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	fixtures := Setup(ctx, t, DefaultTestParams())
	nextBatchIndex := int(fixtures.Params.InitialEpochID.Uint64())
	batchIndexAcceptenceInterval := int(fixtures.Config.BatchIndexAcceptenceInterval)
	t.Run("Future", func(t *testing.T) {
		tx, _ := fixtures.MakeTx(t, 0, nextBatchIndex+batchIndexAcceptenceInterval, 0, 22000)
		err := fixtures.Batcher.EnqueueTx(ctx, tx)
		assert.Error(t, err, ErrBatchIndexTooFarInFuture.Error())
	})
	t.Run("Past", func(t *testing.T) {
		tx, _ := fixtures.MakeTx(t, 0, nextBatchIndex-1, 0, 22000)
		err := fixtures.Batcher.EnqueueTx(ctx, tx)
		assert.Error(t, err, ErrBatchIndexInPast.Error())
	})
	t.Run("Nonce", func(t *testing.T) {
		tx, _ := fixtures.MakeTx(t, 0, nextBatchIndex, 1, 22000)
		err := fixtures.Batcher.EnqueueTx(ctx, tx)
		assert.Error(t, err, ErrNonceMismatch.Error())
	})
	t.Run("Gas", func(t *testing.T) {
		tx, _ := fixtures.MakeTx(t, 0, nextBatchIndex, 0, 2)
		err := fixtures.Batcher.EnqueueTx(ctx, tx)
		assert.ErrorContains(t, err, "tx gas lower than minimum")
	})

	t.Run("ChainID", func(t *testing.T) {
		chainID := fixtures.ChainID
		fixtures.ChainID = big.NewInt(123)
		defer func() { fixtures.ChainID = chainID }()
		tx, _ := fixtures.MakeTx(t, 0, nextBatchIndex+1, 0, 22000)
		err := fixtures.Batcher.EnqueueTx(ctx, tx)
		assert.Error(t, err, ErrWrongChainID.Error())
	})
}

func TestRejectTxNotEnoughFundsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	fixtures := Setup(ctx, t, DefaultTestParams())
	nextBatchIndex := int(fixtures.Params.InitialEpochID.Uint64())
	tx, _ := fixtures.MakeTx(t, 1, nextBatchIndex, 0, 22000)
	err := fixtures.Batcher.EnqueueTx(ctx, tx)
	assert.Error(t, err, ErrCannotPayGasFee.Error())
}

func TestConfirmTransactionsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	fixtures := Setup(ctx, t, DefaultTestParams())
	nextBatchIndex := int(fixtures.Params.InitialEpochID.Uint64())

	for nonce := 0; nonce < 5; nonce++ {
		tx, _ := fixtures.MakeTx(t, 0, nextBatchIndex, nonce, 22000)
		err := fixtures.Batcher.EnqueueTx(ctx, tx)
		assert.NilError(t, err)
	}

	txs, err := fixtures.DB.GetTransactionsByEpoch(ctx, fixtures.Params.InitialEpochID.Bytes())
	assert.NilError(t, err)
	assert.Equal(t, 5, len(txs), "should have exactly one tx: %+v", txs)
	for nonce := 0; nonce < 5; nonce++ {
		assert.Equal(t, cltrdb.TxstatusCommitted, txs[nonce].Status, "expected tx to have status committed: %+v", txs[nonce])
	}
}

func TestCloseBatchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	var err error
	ctx := context.Background()
	fixtures := Setup(ctx, t, DefaultTestParams())
	t.Run("Init", func(t *testing.T) {
		assert.Check(t, fixtures.Batcher.nextBatchChainState != nil, "nextBatchChainState field not initialized")
	})

	t.Run("closeEmptyBatch", func(t *testing.T) {
		// we should be able to successfully close the batch, even though it is empty.
		err = fixtures.Batcher.CloseBatch(ctx)
		assert.NilError(t, err)
		assert.Check(t, fixtures.Batcher.nextBatchChainState == nil, "nextBatchChainState field initialized")
		nextBatchEpoch, _, err := batchhandler.GetNextBatch(ctx, fixtures.DB)
		assert.NilError(t, err)
		assert.Equal(t, nextBatchEpoch.Uint64(), fixtures.Params.InitialEpochID.Uint64()+1)
	})

	t.Run("initChainStateWaitForSequencer", func(t *testing.T) {
		err = fixtures.Batcher.initChainState(ctx)
		assert.Error(t, err, ErrWaitForSequencer.Error())
	})

	t.Run("batchAlreadyExists", func(t *testing.T) {
		fixtures.EthL2Server.SetBatchIndex(fixtures.Params.InitialEpochID.Uint64() + 1)
		err = fixtures.Batcher.initChainState(ctx)
		assert.Error(t, err, ErrBatchAlreadyExists.Error())
	})

	t.Run("initChainStateSetsNextBatchChainState", func(t *testing.T) {
		fixtures.EthL2Server.SetBatchIndex(fixtures.Params.InitialEpochID.Uint64())
		err = fixtures.Batcher.initChainState(ctx)
		assert.NilError(t, err)
		assert.Check(t, fixtures.Batcher.nextBatchChainState != nil, "nextBatchChainState field not initialized")
	})
}

// TestOpenNextBatch checks that we're able to enqueue transactions for the next batch even though
// the l2 chain hasn't processed the last block.
func TestOpenNextBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	var err error
	ctx := context.Background()
	fixtures := Setup(ctx, t, DefaultTestParams())

	err = fixtures.Batcher.CloseBatch(ctx)
	assert.NilError(t, err)
	assert.Check(t, fixtures.Batcher.nextBatchChainState == nil, "nextBatchChainState field initialized")

	nextBatchEpoch, _, err := batchhandler.GetNextBatch(ctx, fixtures.DB)
	assert.NilError(t, err)
	nextBatchIndex := nextBatchEpoch.Uint64()

	assert.Equal(t, nextBatchIndex, fixtures.Params.InitialEpochID.Uint64()+1)
	// we should now be able to enqueue transactions. The batcher however doesn't have
	// information about the current nonce and the balances, so we're able to enqueue
	// transactions that later will be rejected when the l2 chain builds a new block.

	tx, _ := fixtures.MakeTx(t, 0, int(nextBatchIndex), 0, 22000)
	err = fixtures.Batcher.EnqueueTx(ctx, tx)
	assert.NilError(t, err)

	// Account 1 has a zero balance. This transaction will get rejected later.
	tx2, _ := fixtures.MakeTx(t, 1, int(nextBatchIndex), 0, 22000)
	err = fixtures.Batcher.EnqueueTx(ctx, tx2)
	assert.NilError(t, err)

	txs, err := fixtures.DB.GetNonRejectedTransactionsByEpoch(ctx, epochid.Uint64ToEpochID(nextBatchIndex).Bytes())
	assert.NilError(t, err)
	assert.Equal(t, len(txs), 2)
	log.Printf("NoneRejected, not verified: %+v", txs)

	// so, now let's let the l2 chain build a new block
	fixtures.EthL2Server.SetBatchIndex(fixtures.Params.InitialEpochID.Uint64())

	tx3, _ := fixtures.MakeTx(t, 2, int(nextBatchIndex), 0, 22000)
	err = fixtures.Batcher.EnqueueTx(ctx, tx3)
	assert.Error(t, err, ErrCannotPayGasFee.Error())

	assert.Check(t, fixtures.Batcher.nextBatchChainState != nil, "nextBatchChainState field not initialized")

	txs, err = fixtures.DB.GetTransactionsByEpoch(ctx, epochid.Uint64ToEpochID(nextBatchIndex).Bytes())
	assert.NilError(t, err)
	assert.Equal(t, len(txs), 2)

	assert.Equal(t, txs[0].Status, cltrdb.TxstatusCommitted)
	assert.Equal(t, txs[1].Status, cltrdb.TxstatusRejected)
}

// TestDecryptionTrigger tests that closing a batch
// will fill the DB with a DecryptionTrigger that
// includes the hash of the committed transactions hashes.
func TestDecryptionTriggerGeneratedIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	var err error
	ctx := context.Background()
	fixtures := Setup(ctx, t, DefaultTestParams())

	nextBatchEpoch, _, err := batchhandler.GetNextBatch(ctx, fixtures.DB)
	assert.NilError(t, err)
	nextBatchIndex := nextBatchEpoch.Uint64()

	assert.Equal(t, nextBatchIndex, fixtures.Params.InitialEpochID.Uint64())

	tx, txHash := fixtures.MakeTx(t, 0, int(nextBatchIndex), 0, 22000)
	err = fixtures.Batcher.EnqueueTx(ctx, tx)
	assert.NilError(t, err)
	tx2, tx2Hash := fixtures.MakeTx(t, 0, int(nextBatchIndex), 1, 22000)
	err = fixtures.Batcher.EnqueueTx(ctx, tx2)
	assert.NilError(t, err)

	err = fixtures.Batcher.CloseBatch(ctx)
	assert.NilError(t, err)

	triggers, err := fixtures.DB.GetUnsentTriggers(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(triggers), 1)
	trigger := triggers[0]

	expectedHash := shmsg.HashByteList([][]byte{txHash, tx2Hash})
	assert.DeepEqual(t, expectedHash, trigger.BatchHash)

	err = fixtures.DB.UpdateDecryptionTriggerSent(ctx, trigger.EpochID)
	assert.NilError(t, err)

	triggers, err = fixtures.DB.GetUnsentTriggers(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(triggers), 0)
}

func TestDecryptionTriggerInsertOrderingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	var err error
	ctx := context.Background()
	fixtures := Setup(ctx, t, DefaultTestParams())

	trigger1 := cltrdb.InsertTriggerParams{
		EpochID:       epochid.Uint64ToEpochID(2).Bytes(),
		BatchHash:     common.BytesToHash([]byte{1, 0}).Bytes(),
		L1BlockNumber: 666,
	}

	trigger2 := cltrdb.InsertTriggerParams{
		EpochID:       epochid.Uint64ToEpochID(1).Bytes(),
		BatchHash:     common.BytesToHash([]byte{0, 1}).Bytes(),
		L1BlockNumber: 42,
	}
	// insert trigger2 first
	err = fixtures.DB.InsertTrigger(ctx, trigger2)
	assert.NilError(t, err)
	// insert trigger1 second
	err = fixtures.DB.InsertTrigger(ctx, trigger1)
	assert.NilError(t, err)

	triggers, err := fixtures.DB.GetUnsentTriggers(ctx)
	assert.NilError(t, err)

	assert.Equal(t, len(triggers), 2)

	// ordering is by insert order,
	// independent of the epochid or l1blocknumber
	assert.DeepEqual(t, cltrdb.DecryptionTrigger{
		EpochID:       trigger2.EpochID,
		ID:            1,
		BatchHash:     trigger2.BatchHash,
		L1BlockNumber: trigger2.L1BlockNumber,
	}, triggers[0])

	assert.DeepEqual(t, cltrdb.DecryptionTrigger{
		EpochID:       trigger1.EpochID,
		ID:            2,
		BatchHash:     trigger1.BatchHash,
		L1BlockNumber: trigger1.L1BlockNumber,
	}, triggers[1])
}
