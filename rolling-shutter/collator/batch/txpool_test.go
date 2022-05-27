package batch

import (
	"math/big"
	"testing"
	"time"

	txtypes "github.com/shutter-network/txtypes/types"
	"gotest.tools/assert"
)

func TestBatches(t *testing.T) {
	batches := make(Batches, 0)
	// Batches reorders the underlying array
	// on insert
	batches.Insert(2)
	batches.Insert(5)
	batches.Insert(3)
	batches.Insert(1)

	expected := []uint64{1, 2, 3, 5}
	assert.DeepEqual(t, expected, batches.Batches())

	batches.Remove(1)
	batches.Remove(3)

	expected = []uint64{2, 5}
	assert.DeepEqual(t, expected, batches.Batches())

	batches.Insert(4)

	expected = []uint64{2, 4, 5}
	assert.DeepEqual(t, expected, batches.Batches())

	batches.Remove(5)
	batches.Remove(2)
	batches.Remove(4)

	assert.Equal(t, batches.Len(), 0)
}

func TestTxByNonceAndTime(t *testing.T) {
	cfg := newTestConfig(t)
	chainID := big.NewInt(0)
	signer := txtypes.LatestSignerForChainID(chainID)

	makeTx := func(batchIndex, nonce int, tm time.Time) *PendingTransaction {
		// construct a valid transaction
		txData := &txtypes.ShutterTx{
			ChainID:          chainID,
			Nonce:            uint64(nonce),
			GasTipCap:        big.NewInt(10),
			GasFeeCap:        big.NewInt(10),
			Gas:              uint64(100),
			EncryptedPayload: []byte("foo"),
			BatchIndex:       uint64(batchIndex),
		}
		tx, err := txtypes.SignNewTx(cfg.EthereumKey, signer, txData)
		assert.NilError(t, err)

		// marshal tx to bytes
		txBytes, err := tx.MarshalBinary()
		assert.NilError(t, err)
		ptx, err := NewPendingTransaction(signer, txBytes, tm)
		assert.NilError(t, err)
		return ptx
	}

	txs := make(TxByNonceAndTime, 0)
	tm := time.Now()

	tx1 := makeTx(1, 1, tm.AddDate(0, 0, 1))
	tx2 := makeTx(1, 2, tm.AddDate(0, 0, 1))
	tx3 := makeTx(1, 2, tm.AddDate(0, 0, 2))

	txs.Push(tx3)
	txs.Push(tx1)
	txs.Push(tx2)

	// just check that insert order equals pop order
	// the classes' methods can be used by a heap,
	// but the class does not keep sorted order by itself
	assert.DeepEqual(t, txs.Pop().(*PendingTransaction).txBytes, tx3.txBytes)
	assert.DeepEqual(t, txs.Pop().(*PendingTransaction).txBytes, tx1.txBytes)
	assert.DeepEqual(t, txs.Pop().(*PendingTransaction).txBytes, tx2.txBytes)
}

func TestTxPool(t *testing.T) {
	cfg := newTestConfig(t)
	chainID := big.NewInt(0)
	signer := txtypes.LatestSignerForChainID(chainID)

	makeTx := func(batchIndex, nonce int, tm time.Time) *PendingTransaction {
		// construct a valid transaction
		txData := &txtypes.ShutterTx{
			ChainID:          chainID,
			Nonce:            uint64(nonce),
			GasTipCap:        big.NewInt(10),
			GasFeeCap:        big.NewInt(10),
			Gas:              uint64(100),
			EncryptedPayload: []byte("foo"),
			BatchIndex:       uint64(batchIndex),
		}
		tx, err := txtypes.SignNewTx(cfg.EthereumKey, signer, txData)
		assert.NilError(t, err)

		// marshal tx to bytes
		txBytes, err := tx.MarshalBinary()
		assert.NilError(t, err)
		ptx, err := NewPendingTransaction(signer, txBytes, tm)
		assert.NilError(t, err)
		return ptx
	}

	txpool := NewTransactionPool(signer)
	tm := time.Now()

	tx1 := makeTx(1, 1, tm.AddDate(0, 0, 1))
	tx2 := makeTx(2, 2, tm.AddDate(0, 0, 1))
	tx3 := makeTx(2, 2, tm.AddDate(0, 0, 2))

	// insert order should not affect the
	// the sorting order on a later pop
	txpool.Push(tx3)
	txpool.Push(tx1)
	txpool.Push(tx2)

	expected := []uint64{1, 2}
	assert.DeepEqual(t, expected, txpool.Batches().Batches())

	assert.Equal(t, len(txpool.Senders(1)), 1)
	assert.Equal(t, len(txpool.Senders(2)), 1)

	batch1 := txpool.Pop(1)
	batch2 := txpool.Pop(2)
	assert.Equal(t, len(batch1), 1)
	assert.Equal(t, len(batch2), 2)
	// per batch the tx's should now be sorted by (nonce, insert time)
	// independent on insert order
	assert.DeepEqual(t, batch1[0].txBytes, tx1.txBytes)
	assert.DeepEqual(t, batch2[0].txBytes, tx2.txBytes)
	assert.DeepEqual(t, batch2[1].txBytes, tx3.txBytes)

	// The batches should be removed from the pool
	batch1 = txpool.Pop(1)
	batch2 = txpool.Pop(2)
	assert.Equal(t, len(batch1), 0)
	assert.Equal(t, len(batch2), 0)
	// A pop should also remove the contained information from
	// all other sets
	// (sender addresses per batch, set of batch indices)
	assert.Equal(t, len(txpool.Senders(1)), 0)
	assert.Equal(t, len(txpool.Senders(2)), 0)
	assert.Equal(t, len(txpool.Batches().ToUint64s()), 0)
}
