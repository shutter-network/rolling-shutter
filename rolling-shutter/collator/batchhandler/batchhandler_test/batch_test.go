package batchhandler_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	txtypes "github.com/shutter-network/txtypes/types"
	"golang.org/x/sync/errgroup"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/batch"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

func TestStateTransition(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	ctx := context.Background()

	epoch1, _ := epochid.BigToEpochID(common.Big1)
	epoch2, _ := epochid.BigToEpochID(common.Big2)
	params := TestParams{
		GasLimit:       uint64(210000),
		BaseFee:        big.NewInt(1),
		InitialBalance: big.NewInt(1000000),
		TxGasTipCap:    big.NewInt(1),
		TxGasFeeCap:    big.NewInt(2),
		InitialEpochID: epoch1,
	}
	fixtures := Setup(ctx, t, params)
	egroup, ctx := errgroup.WithContext(ctx)

	l2EthClient, err := ethclient.Dial(fixtures.Cfg.SequencerURL)
	assert.NilError(t, err)

	chainID, err := l2EthClient.ChainID(ctx)
	assert.NilError(t, err)
	signer := txtypes.LatestSignerForChainID(chainID)

	l1BlockNumber := uint64(42)
	decryptionKey := []byte("foo")

	epochChan1 := make(chan time.Time)
	epochChan2 := make(chan time.Time)
	var epochTime1, epochTime2 time.Time

	headBatch, err := batch.New(ctx, fixtures.Cfg.InstanceID, epoch1, l1BlockNumber, l2EthClient, nil)
	assert.NilError(t, err)

	// pass in the previous batch to the next batch so it can listen to state-changes
	nextBatch, err := batch.New(ctx, fixtures.Cfg.InstanceID, epoch2, l1BlockNumber, l2EthClient, headBatch)
	assert.NilError(t, err)

	currentTime := time.Now()
	newTx := func(batchIndex, nonce int, receiveTime time.Time) (*transaction.Pending, []byte) {
		tx, txHash := fixtures.MakeTx(batchIndex, nonce, 21000)
		ptx, err := transaction.NewPending(signer, tx, receiveTime)
		assert.NilError(t, err)
		return ptx, txHash
	}
	spamStateTransition := func(b *batch.Batch, epochChan chan<- time.Time, tx, epoch, decryption, confirmation bool) {
		epochID, err := epochid.BigToEpochID(common.Big256)
		assert.NilError(t, err)

		if epoch {
			epochChan <- time.Now()
		}
		if decryption {
			b.DecryptionKey <- []byte("foobar")
		}
		if confirmation {
			b.ConfirmedBatch <- epochID
		}
		if tx {
			ntx, _ := newTx(int(b.Index()), 4, time.Now())
			b.Transaction <- ntx
		}
	}

	// FIXME the nonce / etc based ordering is not used currently in the TransactionQueue
	// when this is implemented, test different insert order for "initial" state
	tx1, tx1Hash := newTx(0, 0, currentTime.Add(time.Second))
	tx2, tx2Hash := newTx(0, 1, currentTime.Add(2*time.Second))
	tx3TooLate, _ := newTx(0, 2, currentTime.Add(3*time.Second))
	tx4invalid, _ := newTx(1, 1, currentTime.Add(4*time.Second))
	tx5, tx5Hash := newTx(1, 2, currentTime.Add(5*time.Second))
	tx6, tx6Hash := newTx(1, 3, currentTime.Add(6*time.Second))

	// This function handles the individual state transitions
	// for the first batch and checks the StateChange
	// transition artifacts (DecryptionTrigger, BatchTx)
	// We specifically step through all states of the batch and
	// manually trigger the events that cause the batch to transition
	// to the next state.
	// In every state, we also trigger all other events that should
	// not effect the state or cause a state transition.
	headBatchHandler := func(ctx context.Context, b *batch.Batch, subscription chan batch.StateChangeResult) error {
		eg, egctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			return headBatch.Run(egctx, epochChan1)
		})

		// the Run() returns with an error, so the errorgroup context is canceled,
		// -> but we don't check for the cancel but only wait to get a subscription value
		stateChange, err := getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.NoState && stateChange.ToState == batch.InitialState) {
			return errors.New("assert failed")
		}
		// Wait for State Pending,
		// In this state we can already include valid transactions in the
		// Batch, and they will get applied immediately to the Batches cached
		// chain-state (balance, nonce) and thus included in
		// the BatchTx.

		// do stuff that shouldn't effect that state
		spamStateTransition(b,
			epochChan1,
			false, // don't spam a transaction
			false, // don't spam an epoch tick
			true,  // spam a decryption key
			true,  // spam a batch confirmation
		)

		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.InitialState && stateChange.ToState == batch.PendingState) {
			return errors.New("assert failed")
		}

		// do stuff that shouldn't effect that state
		spamStateTransition(b,
			epochChan1,
			false, // don't spam a transaction
			false, // don't spam an epoch tick
			true,  // spam a decryption key
			true,  // spam a batch confirmation
		)

		b.Transaction <- tx1
		b.Transaction <- tx2
		// Now progress the epoch,
		// this will cause the batch to transition to State Committed
		epochTime1 = time.Now()
		epochChan1 <- epochTime1

		// State Committed
		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.PendingState && stateChange.ToState == batch.CommittedState) {
			return errors.New("assert failed")
		}
		if err := checkDecryptionTrigger(t, stateChange, fixtures, epoch1, l1BlockNumber, tx1Hash, tx2Hash); err != nil {
			return err
		}

		// do stuff that shouldn't effect that state
		spamStateTransition(b,
			epochChan1,
			true,  // spam a transaction
			true,  // spam an epoch tick
			false, // spam a decryption key
			true,  // spam a batch confirmation
		)
		// shouldn't have any effect in this state (already committed), so we should not see this in
		// the BatchTx later
		b.Transaction <- tx3TooLate
		b.DecryptionKey <- decryptionKey

		// State Decrypted
		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.CommittedState && stateChange.ToState == batch.DecryptedState) {
			return errors.New("assert failed")
		}

		// FIXME re-ordering by gas/insert time not used atm
		if err := checkBatchTx(t, stateChange, chainID,
			decryptionKey,
			epoch1,
			l1BlockNumber,
			tx1.TxBytes, tx2.TxBytes); err != nil {
			return err
		}

		fixtures.EthL2Server.SetNonce(fixtures.Address, 2, "latest")
		fixtures.EthL2Server.SetBalance(fixtures.Address, new(big.Int).SetUint64(1000000-2*(42000)), "latest")

		// do stuff that shouldn't effect that state
		spamStateTransition(b,
			epochChan1,
			true, // spam a transaction
			true, // spam an epoch tick
			true, // spam a decryption key
			true, // spam a batch confirmation, but this is not the one we are waiting for
		)

		b.ConfirmedBatch <- epoch1

		// Wait for state Confirmed
		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.DecryptedState && stateChange.ToState == batch.ConfirmedState) {
			return errors.New("assert failed")
		}

		// do stuff that shouldn't effect that state
		spamStateTransition(b,
			epochChan1,
			true, // spam a transaction
			true, // spam an epoch tick
			true, // spam a decryption key
			true, // spam a batch confirmation
		)

		b.Stop()

		// subscription will get closed after one more transition
		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.ConfirmedState && stateChange.ToState == batch.StoppingState) {
			return errors.New("assert failed")
		}

		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		stateChange.Log().Error().Msg("got this state - head batch")
		if !assert.Check(t, stateChange.FromState == batch.StoppingState && stateChange.ToState == batch.NoState) {
			return errors.New("assert failed")
		}

		t.Log("stopped head batch")
		// the batch was stopped

		// FIXME wrap error maybe
		return eg.Wait()
	}

	// This function handles the individual state transitions
	// for a successive batch and checks the StateChange
	// transition artifacts (DecryptionTrigger, BatchTx)
	// We specifically step through all states of the batch and
	// manually trigger the events that cause the batch to transition
	// to the next state.
	// In every state, we also trigger all other events that should
	// not effect the state or cause a state transition.
	nextBatchHandler := func(ctx context.Context, b *batch.Batch, subscription chan batch.StateChangeResult) error {
		// if the batch is stopped correctly,
		// the subs channel will be closed anyways
		// defer b.Unsubscribe(subscription)
		eg, egctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			return nextBatch.Run(egctx, epochChan2)
		})

		stateChange, err := getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.NoState && stateChange.ToState == batch.InitialState) {
			return errors.New("assert failed")
		}

		// do stuff that shouldn't effect that state
		spamStateTransition(b,
			epochChan2,
			false, // don't spam a transaction
			true,  // spam an epoch tick
			true,  // spam a decryption key
			true,  // spam a batch confirmation
		)

		// This transaction will get added to the "pool" without being validated just now
		// But it will get validated and removed once we transition to state "Pending",
		b.Transaction <- tx4invalid

		// This transaction will get added to the "pool" without being validated
		// Since it is a valid tx, it will get successfully Applied to the batch
		// once we transition to state "Pending"
		b.Transaction <- tx5

		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.InitialState && stateChange.ToState == batch.PendingState) {
			return errors.New("assert failed")
		}
		// do stuff that shouldn't effect that state
		spamStateTransition(b,
			epochChan2,
			false, // don't spam a transaction
			false, // don't spam an epoch tick
			true,  // spam a decryption key
			true,  // spam a batch confirmation
		)
		// This tx will get applied directly
		b.Transaction <- tx6
		// Now progress the epoch,
		// this will cause the batch to transition to State Committed
		epochTime2 = time.Now()
		epochChan2 <- epochTime2

		// State Committed
		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.PendingState && stateChange.ToState == batch.CommittedState) {
			return errors.New("assert failed")
		}
		if !assert.Check(t, len(stateChange.P2PMessages) == 1) {
			return errors.New("assert failed")
		}

		if err := checkDecryptionTrigger(t, stateChange, fixtures, epoch2, l1BlockNumber, tx5Hash, tx6Hash); err != nil {
			return err
		}
		// do stuff that shouldn't effect that state
		spamStateTransition(b,
			epochChan2,
			true,  // spam a transaction
			true,  // spam an epoch tick
			false, // don't spam a decryption key
			true,  // spam a batch confirmation
		)

		b.DecryptionKey <- decryptionKey

		// State Decrypted
		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.CommittedState && stateChange.ToState == batch.DecryptedState) {
			return errors.New("assert failed")
		}

		if err := checkBatchTx(t, stateChange, chainID,
			decryptionKey,
			epoch2,
			l1BlockNumber,
			tx5.TxBytes, tx6.TxBytes); err != nil {
			return err
		}

		fixtures.EthL2Server.SetNonce(fixtures.Address, 4, "latest")
		fixtures.EthL2Server.SetBalance(fixtures.Address, new(big.Int).SetUint64(1000000-4*(42000)), "latest")

		// do stuff that shouldn't effect that state
		spamStateTransition(b,
			epochChan2,
			true, // spam a transaction
			true, // spam an epoch tick
			true, // spam a decryption key
			true, // spam a batch confirmation, but not the one we are waiting for
		)
		b.ConfirmedBatch <- epoch2

		// Wait for state Confirmed
		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.DecryptedState && stateChange.ToState == batch.ConfirmedState) {
			return errors.New("assert failed")
		}
		spamStateTransition(b,
			epochChan2,
			true, // spam a transaction
			true, // spam an epoch tick
			true, // spam a decryption key
			true, // spam a batch confirmation
		)

		b.Stop()

		// subscription will get closed after one more transition
		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		if !assert.Check(t, stateChange.FromState == batch.ConfirmedState && stateChange.ToState == batch.StoppingState) {
			return errors.New("assert failed")
		}

		stateChange, err = getStateChange(egctx, subscription, false)
		if err != nil {
			return err
		}
		stateChange.Log().Error().Msg("got this state - next batch")
		if !assert.Check(t, stateChange.FromState == batch.StoppingState && stateChange.ToState == batch.NoState) {
			return errors.New("assert failed")
		}
		// the batch was stopped
		t.Log("stopped next batch")

		// FIXME wrap err maybe
		return eg.Wait()
	}

	// make sure the broker have the subcription
	// channels registered before the broker is started.
	// that way we won't miss the first state-change
	headBatchSubscription := headBatch.Broker.Subscribe(1)
	nextBatchSubscription := nextBatch.Broker.Subscribe(1)
	// run our custom test state-handlers,
	// they will block until the subscription
	// to the state changes is finished
	egroup.Go(func() error {
		return headBatchHandler(ctx, headBatch, headBatchSubscription)
	})
	egroup.Go(func() error {
		return nextBatchHandler(ctx, nextBatch, nextBatchSubscription)
	})

	err = egroup.Wait()
	assert.NilError(t, err)

	assert.Assert(t, epochTime1.Before(epochTime2))
	assert.Check(t, (<-tx1.Result).Success)
	assert.Check(t, (<-tx2.Result).Success)
	assert.Check(t, !(<-tx3TooLate.Result).Success)
	assert.Check(t, !(<-tx4invalid.Result).Success)
	assert.Check(t, (<-tx5.Result).Success)
	assert.Check(t, (<-tx6.Result).Success)
}

func checkDecryptionTrigger(t *testing.T,
	stateChange batch.StateChangeResult,
	fixtures *Fixture,
	epochID epochid.EpochID,
	blockNumber uint64,
	txHashes ...[]byte,
) error {
	t.Helper()
	if !assert.Check(t, len(stateChange.P2PMessages) == 1) {
		return errors.New("assert failed")
	}

	trigger := stateChange.P2PMessages[0].(*shmsg.DecryptionTrigger)
	expectedTxHash := shmsg.HashTransactions(txHashes)
	expectedTrigger := &shmsg.DecryptionTrigger{
		InstanceID:       fixtures.Cfg.InstanceID,
		EpochID:          epochID.Bytes(),
		BlockNumber:      blockNumber,
		TransactionsHash: expectedTxHash,
	}

	if !assert.Check(t, gocmp.Equal(
		trigger,
		expectedTrigger,
		gocmp.Comparer(compareByte),
		cmpopts.IgnoreFields(shmsg.DecryptionTrigger{}, "Signature"),
		cmpopts.IgnoreUnexported(shmsg.DecryptionTrigger{}),
	)) {
		return errors.New("assert failed")
	}
	return nil
}

func checkBatchTx(t *testing.T,
	stateChange batch.StateChangeResult,
	chainID *big.Int,
	decryptionKey []byte,
	epochID epochid.EpochID,
	blockNumber uint64,
	transactions ...[]byte,
) error {
	t.Helper()
	if !assert.Check(t, len(stateChange.SequencerTransactions) == 1) {
		return errors.New("assert failed")
	}

	batchTx := stateChange.SequencerTransactions[0].(*txtypes.BatchTx)

	expectedBatchTx := &txtypes.BatchTx{
		ChainID:       chainID,
		DecryptionKey: decryptionKey,
		BatchIndex:    epochID.Uint64(),
		L1BlockNumber: blockNumber,
		Transactions:  transactions,
	}

	if !assert.Check(t,
		gocmp.Equal(
			batchTx,
			expectedBatchTx,
			gocmp.Comparer(compareBigInt),
			gocmp.Comparer(compareByte),
			// we don't know when the struct should have been created,
			// so we can't compare
			cmpopts.IgnoreFields(txtypes.BatchTx{}, "Timestamp"))) {
		return errors.New("assert failed")
	}
	return nil
}

func getStateChange(
	ctx context.Context,
	subscription chan batch.StateChangeResult,
	expectClosedChan bool,
) (batch.StateChangeResult, error) {
	select {
	case stateChange, ok := <-subscription:
		if expectClosedChan == ok {
			return batch.StateChangeResult{}, errors.New("expectClosedChan parameter didn't meet expectations")
		}
		return stateChange, nil
	case <-ctx.Done():
		err := errors.Wrap(ctx.Err(), "batch Run context got canceled")
		return batch.StateChangeResult{}, err
	}
}
