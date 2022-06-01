package batch

import (
	"bytes"
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/jackc/pgx/v4"
	txtypes "github.com/shutter-network/txtypes/types"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/cltrdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testdb"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func newTestConfig(t *testing.T) config.Config {
	t.Helper()

	ethereumKey, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	return config.Config{
		EthereumURL:         "http://127.0.0.1:8454",
		SequencerURL:        "http://127.0.0.1:8455",
		EthereumKey:         ethereumKey,
		ExecutionBlockDelay: uint32(5),
		InstanceID:          123,
	}
}

type testParams struct {
	gasLimit       uint64
	initialBalance *big.Int
	baseFee        *big.Int
	txGasTipCap    *big.Int
	txGasFeeCap    *big.Int
	initialEpochID epochid.EpochID
}

type fixture struct {
	cfg          config.Config
	ethServer    *MockEthServer
	batchHandler *BatchHandler
	makeTx       func(batchIndex, nonce, gas int) ([]byte, []byte)
	address      common.Address
	coinbase     common.Address
}

func setup(ctx context.Context, t *testing.T, params testParams) *fixture {
	t.Helper()

	cfg := newTestConfig(t)
	eth := RunMockEthServer(t)
	cfg.SequencerURL = eth.URL

	db, dbpool, dbteardown := testdb.NewCollatorTestDB(ctx, t)
	t.Cleanup(eth.Teardown)
	t.Cleanup(dbteardown)

	address := ethcrypto.PubkeyToAddress(cfg.EthereumKey.PublicKey)
	chainID := big.NewInt(0)
	gasLimit := params.gasLimit
	signer := txtypes.LatestSignerForChainID(chainID)
	coinbase := common.HexToAddress("0x0000000000000000000000000000000000000000")

	// Set the values on the dummy rpc server
	eth.SetBalance(address, params.initialBalance, "latest")
	eth.SetBalance(coinbase, big.NewInt(0), "latest")
	eth.SetNonce(address, uint64(0), "latest")
	eth.SetChainID(chainID)
	eth.SetBlock(params.baseFee, gasLimit, "latest")

	// set initial ("next") epoch id manually,
	// this is usually done in the collator and not in the handler
	err := db.SetNextEpochID(ctx, params.initialEpochID.Bytes())
	assert.NilError(t, err)

	// New batch handler, this will already query the eth-server
	bh, err := NewBatchHandler(cfg, dbpool)
	assert.NilError(t, err)
	assert.Equal(t, bh.LatestEpochID(), params.initialEpochID)

	makeTx := func(batchIndex, nonce, gas int) ([]byte, []byte) {
		// construct a valid transaction
		txData := &txtypes.ShutterTx{
			ChainID:          chainID,
			Nonce:            uint64(nonce),
			GasTipCap:        params.txGasTipCap,
			GasFeeCap:        params.txGasFeeCap,
			Gas:              uint64(gas),
			EncryptedPayload: []byte("foo"),
			BatchIndex:       uint64(batchIndex),
		}
		tx, err := txtypes.SignNewTx(cfg.EthereumKey, signer, txData)
		assert.NilError(t, err)

		// marshal tx to bytes
		txBytes, err := tx.MarshalBinary()
		assert.NilError(t, err)
		return txBytes, tx.Hash().Bytes()
	}

	return &fixture{
		cfg:          cfg,
		ethServer:    eth,
		batchHandler: bh,
		makeTx:       makeTx,
		address:      address,
		coinbase:     coinbase,
	}
}

func TestHandlerStateProgression(t *testing.T) {
	ctx := context.Background()
	epoch1, _ := epochid.BigToEpochID(common.Big1)
	epoch2, _ := epochid.BigToEpochID(new(big.Int).Add(epoch1.Big(), common.Big1))
	params := testParams{
		gasLimit:       uint64(210000),
		baseFee:        big.NewInt(1),
		initialBalance: big.NewInt(210000),
		txGasTipCap:    big.NewInt(1),
		txGasFeeCap:    big.NewInt(2),
		initialEpochID: epoch1,
	}
	fixtures := setup(ctx, t, params)

	// enqueue a batch 2 tx
	// check this should not be processed already, only enqueued for later
	// in the txpool
	tx3, _ := fixtures.makeTx(2, 2, 21000)
	err := fixtures.batchHandler.EnqueueTx(ctx, tx3)
	assert.NilError(t, err)
	assertTransaction(t, fixtures.batchHandler, tx3, epoch2, 0)

	// enqueue a batch 1 tx, this goes to the "Batch" state directly,
	// since we already finalized the 0,0 epoch
	tx1, tx1Hash := fixtures.makeTx(1, 0, 21000)
	err = fixtures.batchHandler.EnqueueTx(ctx, tx1)
	assert.NilError(t, err)
	assertTransaction(t, fixtures.batchHandler, tx1, epoch1, 0)
	// enqueue another batch 1 tx
	tx2, tx2Hash := fixtures.makeTx(1, 1, 21000)
	err = fixtures.batchHandler.EnqueueTx(ctx, tx2)
	assert.NilError(t, err)
	assertTransaction(t, fixtures.batchHandler, tx2, epoch1, 1)

	// check there should be 2 tx in the current batch
	assert.Equal(t, fixtures.batchHandler.LatestEpochID(), epoch1)
	assert.Equal(t, fixtures.batchHandler.latestBatch.Transactions().Len(), 2)
	assert.DeepEqual(t, fixtures.batchHandler.txpool.Batches().ToUint64s(), []uint64{2})

	// batch 2
	msgs, err := fixtures.batchHandler.StartNextEpoch(ctx, uint32(1))
	assert.NilError(t, err)

	// make sure decryption trigger is stored in db
	assertDecryptionTrigger(t, fixtures.cfg, fixtures.batchHandler, msgs, [][]byte{tx1Hash, tx2Hash}, epoch1)

	// The new batch still doesn't exist yet when we haven't processed
	// the decryption key of the previous one yet
	assert.Equal(t, fixtures.batchHandler.LatestEpochID(), epoch1)
	assert.Equal(t, fixtures.batchHandler.latestBatch.Transactions().Len(), 2)

	// insert another tx in this batch, but not queued
	tx4, _ := fixtures.makeTx(2, 3, 21000)
	err = fixtures.batchHandler.EnqueueTx(ctx, tx4)
	assert.NilError(t, err)
	assertTransaction(t, fixtures.batchHandler, tx4, epoch2, 1)

	// Register the state update on the next incoming tx
	hookFunc := func(me *MockEthServer, tx *txtypes.Transaction) bool {
		assert.Equal(t, tx.Type(), uint8(txtypes.BatchTxType))
		assert.Equal(t, len(tx.Transactions()), 2)
		assert.Equal(t, tx.BatchIndex(), uint64(1))

		t.Logf("addr: %s, coinbase: %s", fixtures.address.Hex(), fixtures.coinbase.Hex())
		me.SetBalance(fixtures.address, big.NewInt(126000), "latest")
		me.SetBalance(fixtures.coinbase, big.NewInt(42000), "latest")
		me.SetNonce(fixtures.address, uint64(2), "latest")
		// return true to only execute the hook once on
		// the next transaction
		return true
	}
	fixtures.ethServer.RegisterTxHook(hookFunc)

	// batch 1 decryption key
	// this sends a BatchTx that triggers the registered hook function
	mess, err := fixtures.batchHandler.HandleDecryptionKey(ctx, epoch1, []byte("key1"))
	assert.Check(t, mess == nil)
	assert.NilError(t, err)

	// Now that we handled the decryption key,
	// the (2,1) batch should exist and
	// the transactions from the txpool should be applied
	// to it already
	assert.Equal(t, fixtures.batchHandler.LatestEpochID(), epoch2)
	assert.Equal(t, fixtures.batchHandler.latestBatch.Transactions().Len(), 2)
	assert.DeepEqual(t, fixtures.batchHandler.txpool.Batches().Len(), 0)

	// insert another tx in this batch
	tx5, _ := fixtures.makeTx(2, 4, 21000)
	err = fixtures.batchHandler.EnqueueTx(ctx, tx5)
	assert.NilError(t, err)
	assertTransaction(t, fixtures.batchHandler, tx5, epoch2, 2)

	// this should not be queued but immediately be
	// processed in the current batch state
	assert.Equal(t, fixtures.batchHandler.txpool.Batches().Len(), 0)
	assert.Equal(t, fixtures.batchHandler.latestBatch.Transactions().Len(), 3)
}

func TestHandlerFailingValidation(t *testing.T) {
	ctx := context.Background()
	epoch1, _ := epochid.BigToEpochID(common.Big0)
	epoch2, _ := epochid.BigToEpochID(new(big.Int).Add(epoch1.Big(), common.Big1))
	params := testParams{
		gasLimit:       uint64(210000),
		baseFee:        big.NewInt(1),
		initialBalance: big.NewInt(420002),
		txGasTipCap:    big.NewInt(1),
		txGasFeeCap:    big.NewInt(2),
		initialEpochID: epoch1,
	}
	fixtures := setup(ctx, t, params)

	// gas amount is lower than minimum (21000)
	txBytes, _ := fixtures.makeTx(0, 0, 21000)
	// mangle up the tx data
	txBytes[0] = 0xe
	txBytes[1] = 0xf
	err := fixtures.batchHandler.EnqueueTx(ctx, txBytes)
	assert.ErrorContains(t, err, "can't decode transaction bytes")

	// gas amount is lower than minimum (21000)
	txBytes, _ = fixtures.makeTx(0, 0, 20999)
	err = fixtures.batchHandler.EnqueueTx(ctx, txBytes)
	assert.ErrorContains(t, err, "tx gas lower than minimum")

	// enqueue initial valid tx
	txBytes, _ = fixtures.makeTx(0, 0, 189000)
	err = fixtures.batchHandler.EnqueueTx(ctx, txBytes)
	assert.NilError(t, err)

	// this is above the gas-minimum, but it results
	// in the gas limit overflowing
	txBytes, _ = fixtures.makeTx(0, 1, 21001)
	err = fixtures.batchHandler.EnqueueTx(ctx, txBytes)
	assert.ErrorContains(t, err, "gas limit reached")

	// this is fine gas-wise, but does not increment the nonce
	txBytes, _ = fixtures.makeTx(0, 0, 21000)
	err = fixtures.batchHandler.EnqueueTx(ctx, txBytes)
	assert.ErrorContains(t, err, "nonce mismatch, want: 1,got: 0")

	// don't allow transactions too far in the future
	// current batch + 5 is the last one allowed
	txBytes, _ = fixtures.makeTx(5, 0, 21000)
	err = fixtures.batchHandler.EnqueueTx(ctx, txBytes)
	assert.NilError(t, err)
	txBytes, _ = fixtures.makeTx(6, 0, 21000)
	err = fixtures.batchHandler.EnqueueTx(ctx, txBytes)
	assert.ErrorContains(t, err, "batch too far in the future")

	// batch 1
	_, err = fixtures.batchHandler.StartNextEpoch(ctx, uint32(1))
	assert.NilError(t, err)

	// don't allow transactions for the last epoch,
	// once we started the next one.
	// This has to be the case even when the latestBatch has
	// not been updated yet (no decryption key received)
	txBytes, _ = fixtures.makeTx(0, 0, 21000)
	err = fixtures.batchHandler.EnqueueTx(ctx, txBytes)
	assert.ErrorContains(t, err, "historic batch index")

	// batch 0 decryption key
	mess, err := fixtures.batchHandler.HandleDecryptionKey(ctx, epoch1, []byte("key1"))
	assert.Check(t, mess == nil)
	assert.NilError(t, err)
	assert.Equal(t, fixtures.batchHandler.LatestEpochID(), epoch2)
}

func assertTransaction(t *testing.T, bh *BatchHandler, txBytes []byte, epochID epochid.EpochID, expectedIndex int) {
	t.Helper()
	ctx := context.Background()

	foundIndex := -1
	err := bh.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
		db := cltrdb.New(dbtx)
		txs, err := db.GetTransactionsByEpoch(ctx, epochID.Bytes())
		assert.NilError(t, err)
		for i, tx := range txs {
			if cmp.DeepEqual(tx, txBytes)().Success() {
				foundIndex = i
				break
			}
		}
		return nil
	})
	assert.NilError(t, err)
	assert.Equal(t, foundIndex, expectedIndex)
}

func assertDecryptionTrigger(t *testing.T,
	cfg config.Config,
	bh *BatchHandler,
	msgs []shmsg.P2PMessage,
	txHashes [][]byte,
	epochID epochid.EpochID,
) {
	t.Helper()
	ctx := context.Background()

	transactionsHash := shmsg.HashTransactions(txHashes)
	// make sure decryption trigger is stored in db
	err := bh.dbpool.BeginFunc(ctx, func(dbtx pgx.Tx) error {
		db := cltrdb.New(dbtx)
		stored, err := db.GetTrigger(ctx, epochID.Bytes())
		assert.NilError(t, err)
		assert.Check(t, bytes.Equal(stored.EpochID, epochID.Bytes()))
		assert.DeepEqual(t, stored.BatchHash, transactionsHash)
		return nil
	})
	assert.NilError(t, err)

	// make sure output is trigger message
	assert.Equal(t, len(msgs), 1)
	triggerMsg := msgs[0].(*shmsg.DecryptionTrigger)
	assert.Equal(t, triggerMsg.InstanceID, cfg.InstanceID)
	assert.Check(t, bytes.Equal(triggerMsg.EpochID, epochID.Bytes()))
	assert.DeepEqual(t, triggerMsg.TransactionsHash, transactionsHash)
	address := ethcrypto.PubkeyToAddress(cfg.EthereumKey.PublicKey)
	signatureCorrect, err := shmsg.VerifySignature(triggerMsg, address)
	assert.NilError(t, err)
	assert.Check(t, signatureCorrect)
}
