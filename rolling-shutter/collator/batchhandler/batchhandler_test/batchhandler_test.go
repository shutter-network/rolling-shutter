package batchhandler_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gocmp "github.com/google/go-cmp/cmp"
	txtypes "github.com/shutter-network/txtypes/types"
	"golang.org/x/sync/errgroup"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/sequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testlog"
)

func init() {
	testlog.Setup()
}

// `TestBatchHandler` spawns a BatchHandler in conjunction with a
// `p2pMessagingMock` and a mocked sequencer server (`MockEthServer`).
// The BatchHandler then receives different user-transactions
// during different epochs.
// The coordination of spawning the user-transactions and ending
// the test conducted in a registered hook-function in the MockEthServer,
// that is called everytime the MockEthServer receives a BatchTx via the
// `eth_sendRawTransaction` JSON-RPC endpoint.
//
//nolint:funlen //don't fix this, BatchHandler is deprecated and will get removed
func TestBatchHandlerIntegration(t *testing.T) {
	t.Skip("BatchHandler is deprecated and will get removed")

	ctx := context.Background()

	epoch1, _ := epochid.BigToEpochID(common.Big1)
	params := TestParams{
		GasLimit:       uint64(210000),
		BaseFee:        big.NewInt(1),
		InitialBalance: big.NewInt(1000000),
		TxGasTipCap:    big.NewInt(1),
		TxGasFeeCap:    big.NewInt(2),
		InitialEpochID: epoch1,
		EpochDuration:  2 * time.Second,
	}
	dbctx := context.Background()
	fixtures := Setup(dbctx, t, params)
	done := make(chan error)

	userCtx := context.Background()
	var txBytes1, txBytes2, txBytes3, txBytes4, txBytes5, txBytes6, txBytes7, txBytes8, txBytes9, txBytes10 []byte

	// create the result promises
	tx1Result := make(chan transaction.Result, 1)
	tx2Result := make(chan transaction.Result, 1)
	tx3Result := make(chan transaction.Result, 1)
	tx4Result := make(chan transaction.Result, 1)
	tx5Result := make(chan transaction.Result, 1)
	tx6Result := make(chan transaction.Result, 1)
	tx7Result := make(chan transaction.Result, 1)
	tx8Result := make(chan transaction.Result, 1)
	tx9Result := make(chan transaction.Result, 1)
	tx10Result := make(chan transaction.Result, 1)
	currentBatchIndex := uint64(0)

	failFunc := func(err error) {
		done <- err
		close(done)
	}

	fixtures.EthL2Server.RegisterTxHook(func(me *sequencer.MockEthServer, tx *txtypes.Transaction) bool {
		assertEqual(t, failFunc, tx.Type(), uint8(txtypes.BatchTxType))
		assertEqual(t, failFunc, tx.BatchIndex(), currentBatchIndex+1)
		currentBatchIndex++

		epoch := epochid.Uint64ToEpochID(tx.BatchIndex())
		// trigger the batch-confirmation handler in the batch-handler directly.
		// we don't have an endpoint for batch-confirmation yet in the sequencer (and mock-sequencer)

		defer func() {
			fixtures.BatchHandler.ConfirmedBatch() <- epoch
		}()

		switch tx.BatchIndex() {
		case 1:
			assertEqual(t, failFunc, len(tx.Transactions()), 0)
			// valid, applied in batch 2
			txBytes1, _ = fixtures.MakeTx(2, 0, 22000)
			res1 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes1)
			go func() {
				tx1Result <- <-res1
			}()
			// valid, applied in batch 4
			txBytes2, _ = fixtures.MakeTx(4, 1, 22000)
			res2 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes2)
			go func() {
				tx2Result <- <-res2
			}()

			// invalid (decreasing nonce), applied in batch 4
			txBytes3, _ = fixtures.MakeTx(4, 0, 22000)
			res3 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes3)
			go func() {
				// goes to the pool but will then be dropped due to invalid nonce
				tx3Result <- <-res3
			}()

			// invalid (non-monotonically increasing nonce), applied in batch 4
			txBytes10, _ = fixtures.MakeTx(4, 3, 22000)
			res10 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes10)
			go func() {
				// goes to the pool but will then be dropped due to invalid nonce
				tx10Result <- <-res10
			}()

			// valid, applied in the batch that is just on the edge of too far in the future
			txBytes9, _ = fixtures.MakeTx(batchhandler.SizeBatchPool+1, 3, 22000)
			res9 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes9)
			go func() {
				tx9Result <- <-res9
			}()

		case 2:
			assertEqual(
				t,
				failFunc,
				tx.Transactions(),
				[][]byte{txBytes1},
				gocmp.Comparer(compareByte),
			)
			// with above test-parameters, a tx costs 2xGas of the tx
			// and the node get's 1xGas
			me.SetBalance(fixtures.Address, big.NewInt(1000000-44000), "latest")
			me.SetBalance(fixtures.Coinbase, big.NewInt(22000), "latest")
			me.SetNonce(fixtures.Address, uint64(1), "latest")
		case 3:
			assertEqual(t, failFunc, len(tx.Transactions()), 0)

			// valid, applied in batch 4
			txBytes4, _ = fixtures.MakeTx(4, 2, 22000)
			res4 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes4)
			go func() {
				tx4Result <- <-res4
			}()

			// invalid (higher than gas limit), applied in batch 4
			txBytes5, _ = fixtures.MakeTx(4, 3, 210001-44000)
			res5 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes5)
			go func() {
				tx5Result <- <-res5
			}()

			// invalid, batch too far in the future doesn't exist in the batch-pool
			txBytes6, _ = fixtures.MakeTx(3+batchhandler.SizeBatchPool+1, 3, 22000)
			res6 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes6)
			go func() {
				tx6Result <- <-res6
			}()

		case 4:
			assertEqual(
				t,
				failFunc,
				tx.Transactions(),
				[][]byte{txBytes2, txBytes4},
				gocmp.Comparer(compareByte),
			)
			me.SetBalance(fixtures.Address, big.NewInt(1000000-(3*44000)), "latest")
			me.SetBalance(fixtures.Coinbase, big.NewInt(3*22000), "latest")
			me.SetNonce(fixtures.Address, uint64(3), "latest")

			// invalid, batch in the past doesn't exist in the batch pool
			// or has a state that doesn't accept transactions
			txBytes7, _ = fixtures.MakeTx(3, 3, 22000)
			res7 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes7)
			go func() {
				tx7Result <- <-res7
			}()

			// invalid (lower than gas-minimum), applied in batch 5
			txBytes8, _ = fixtures.MakeTx(5, 3, 20999)
			res8 := fixtures.BatchHandler.EnqueueTx(userCtx, txBytes8)
			go func() {
				tx8Result <- <-res8
			}()

		case batchhandler.SizeBatchPool + 1:
			assertEqual(
				t,
				failFunc,
				tx.Transactions(),
				[][]byte{txBytes9},
				gocmp.Comparer(compareByte),
			)
			me.SetBalance(fixtures.Address, big.NewInt(1000000-(4*44000)), "latest")
			me.SetBalance(fixtures.Coinbase, big.NewInt(4*22000), "latest")
			me.SetNonce(fixtures.Address, uint64(4), "latest")
			assertEqual(t, failFunc, len(tx.Transactions()), 1)
		case batchhandler.SizeBatchPool + 2:
			// FIXME this is closed in the test at batch 7, and then somehow the batchhandler goes on..
			// -> this causes batch 8 to never receive the  batch inclusion confirmation
			close(done)
			return true
		default:
			// in all other batches, we don't expect user transactions
			// either because they were invalid, or because we didn't send any
			assertEqual(t, failFunc, len(tx.Transactions()), 0)
		}
		// return false will not deregister the hook
		// after the first execution
		return false
	})

	eg, errctx := errgroup.WithContext(ctx)

	p2pctx, p2pcancel := context.WithCancel(errctx)

	eg.Go(func() error {
		return p2pMessagingMock(p2pctx, t, failFunc, fixtures.Cfg, fixtures.BatchHandler)
	})
	eg.Go(func() error {
		return fixtures.BatchHandler.Run(errctx)
	})

	eg.Go(func() error {
		defer p2pcancel()
		defer fixtures.BatchHandler.Stop()

		select {
		case err := <-done:
			if err == nil {
				return nil
			}
			return err
		case <-errctx.Done():
			return errctx.Err()
		}
	})

	err := eg.Wait()
	assert.NilError(t, err)

	assert.Check(t, ((<-tx1Result).Success))
	assert.Check(t, ((<-tx2Result).Success))
	assert.Check(t, !((<-tx3Result).Success))
	assert.Check(t, ((<-tx4Result).Success))
	assert.Check(t, !((<-tx5Result).Success))
	assert.Check(t, !((<-tx6Result).Success))
	assert.Check(t, !((<-tx7Result).Success))
	assert.Check(t, !((<-tx8Result).Success))
	assert.Check(t, ((<-tx9Result).Success))
	assert.Check(t, !((<-tx10Result).Success))
}
