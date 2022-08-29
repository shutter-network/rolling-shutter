package batch

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/sequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shmsg"
)

const (
	minimumTxGas   uint64 = 21000
	BatchSizeLimit int    = 8 * 1024
)

const (
	NoState StateEnum = iota
	InitialState
	PendingState
	CommittedState
	DecryptedState
	ConfirmedState
	StoppingState
)

type StateEnum int

func (s StateEnum) String() string {
	switch s {
	case NoState:
		return "nostate"
	case InitialState:
		return "initial"
	case PendingState:
		return "pending"
	case CommittedState:
		return "committed"
	case DecryptedState:
		return "decrypted"
	case ConfirmedState:
		return "confirmed"
	case StoppingState:
		return "stopping"
	}
	return ""
}

type StateChangeError struct {
	Err error
}

type StateChangeResult struct {
	EpochID               epochid.EpochID
	FromState             StateEnum
	ToState               StateEnum
	P2PMessages           []shmsg.P2PMessage
	SequencerTransactions []txtypes.TxData
	Errors                []StateChangeError
}

func (st *StateChangeResult) Log() *zerolog.Logger {
	logger := log.With().
		Dict("StateTransition", zerolog.Dict().
			// TODO the reflection based encoding is not
			// consistent with the other logging (e.g. no hex encoding)
			Interface("messages", st.P2PMessages).
			Interface("transactions", st.SequencerTransactions).
			Interface("errors", st.Errors).
			Str("fromState", st.FromState.String()).
			Str("toState", st.ToState.String()),
		).
		Logger()
	return &logger
}

// `State` contains the state-transition logic of a batch on how to react to different
// input-events during it's lifetime.
//
// For all methods of `State` that return a `State` object themselves,
// the following holds true:
// If the return value is the same as the `State` object it has been called
// on, no state transition will be conducted.
// If the return value is a different from the `State` object it has been called
// on, no state transition will be conducted.
type State interface {
	// `StateEnum` returns the enum associated with that state.
	// There has to be a 1:1 relationship from StateEnum <-> State.
	StateEnum() StateEnum

	// `Process` defines the actions to be taken as an immediate effect of
	// the state-transition. This is the first method to be
	// called after a state-transition and the resulting
	// `StateChangeResult` will be emitted to all subscribed observers.
	Process(*Batch) *StateChangeResult

	// `Post` is called immediately after the state-change has been
	// processed.
	Post(*Batch) State

	// `OnEpochTick` defines the actions to be taken as an immediate
	// effect of a batch receiving an epoch tick.
	// The second argument is the time value of the tick.
	OnEpochTick(*Batch, time.Time) State

	// `OnDecryptionKey` defines the actions to be taken as an immediate
	// effect of a batch receiving an decryption key specificly dedicated
	// for that batch.
	// The second argument is the byte-encoded decryption-key.
	OnDecryptionKey(*Batch, []byte) State

	// `OnTransaction` defines the actions to be taken as an immediate
	// effect of a batch receiving an epoch tick.
	// The second argument is the users `Pending` transaction.
	OnTransaction(*Batch, *transaction.Pending) State

	// `OnTransaction` defines the actions to be taken as an immediate
	// effect of a batch receiving a stop signal.
	OnStop(*Batch) State

	// `OnStateChangePrevious` defines the actions to be taken as an immediate
	// effect of the previous batch emitting a state-change resulting
	// from that batches `State`'s `Process` method.
	// The second argument is that `States` emitted `StateChangeResult`
	OnStateChangePrevious(batch *Batch, stateChange StateChangeResult) State

	// `OnBatchConfirmation` defines the actions to be taken as an immediate
	// effect of a batch receiving a newly confirmed batch on the rollup.
	// The second argument is the confirmed batches epoch-id.
	OnBatchConfirmation(batch *Batch, epochID epochid.EpochID) State
}

func ValidateGasParams(tx *txtypes.Transaction, baseFee *big.Int) error {
	if tx.Gas() < minimumTxGas {
		return errors.Errorf("tx gas lower than minimum (%v < %v)", tx.Gas(), minimumTxGas)
	}
	if tx.GasFeeCap().Cmp(tx.GasTipCap()) < 0 {
		return errors.Errorf("gas fee cap lower than gas tip cap (%v < %v)", tx.GasFeeCap(), tx.GasTipCap())
	}
	if tx.GasFeeCap().Cmp(baseFee) < 0 {
		return errors.Errorf("gas fee cap lower than header base fee (%v < %v)", tx.GasFeeCap(), baseFee)
	}
	return nil
}

// CalculateGasCost calculates the overall gas-cost to be deducted from the transaction
// sender's account balance in order for the transaction to be included in the batch.
// The deduction will be applied in the sequencers state transition function.
// The collator has to calculate this prior to including the transaction in the batch-proposal
// in order to check that the sender of the transaction has enough funds to pay for
// processing of the transaction.
// CalculateGasCost should only be called on a Transaction that has been validated
// with ValidateGasParams.
func CalculateGasCost(tx *txtypes.Transaction, baseFee *big.Int) *big.Int {
	priorityFeeGasPrice := math.BigMin(tx.GasTipCap(), new(big.Int).Sub(tx.GasFeeCap(), baseFee))
	gasPrice := new(big.Int).Add(priorityFeeGasPrice, baseFee)
	return new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(tx.Gas()))
}

// CalculatePriorityFee calculates the part of the overall gas-cost to be deducted from the transaction
// sender's account balance and added to the collators (coinbase) account.
// The deduction will be applied in the sequencers state transition function.
// The collator can calculate this prior to including the transaction in the batch-proposal
// in order to decide which transactions should be included in the batch based on the
// value added to the collators balance.
// CalculatePriorityFee should only be called on a Transaction that has been validated
// with ValidateGasParams.
func CalculatePriorityFee(tx *txtypes.Transaction, baseFee *big.Int) *big.Int {
	priorityFeeGasPrice := math.BigMin(tx.GasTipCap(), new(big.Int).Sub(tx.GasFeeCap(), baseFee))
	return new(big.Int).Mul(priorityFeeGasPrice, new(big.Int).SetUint64(tx.Gas()))
}

func New(
	ctx context.Context, instanceID uint64, epochID epochid.EpochID, l1BlockNumber uint64, client *ethclient.Client,
	previousBatch *Batch,
) (*Batch, error) {
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	// batchindex not necessarily the same as the l2blocknumber.
	// just query for the current state of the addresses balance/nonce.
	// since the collator is the only one progressing the balance/nonce state,
	// this is fine as long as the current batch is not submitted
	// number=nil means the latest state from the node
	// FIXME this might not work anymore with the async Batch creation model
	block, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	b := &Batch{
		mu:             sync.RWMutex{},
		previous:       previousBatch,
		Broker:         medley.StartNewBroker[StateChangeResult](false),
		instanceID:     instanceID,
		epochID:        epochID,
		L1BlockNumber:  l1BlockNumber,
		ChainID:        chainID,
		decryptionKey:  []byte{},
		gasPool:        0,
		block:          block,
		chainState:     sequencer.NewCached(client, nil),
		committedTxs:   transaction.NewQueue(),
		txPool:         transaction.NewQueue(),
		ConfirmedBatch: make(chan epochid.EpochID),
		DecryptionKey:  make(chan []byte),
		Transaction:    make(chan *transaction.Pending),
		stopSignal:     make(chan struct{}),
		stoppedResult:  make(chan error, 1),
		// this channel should be deactivated
		// initially
		subscription: nil,
	}
	b.gasPool.AddGas(block.GasLimit())
	return b, nil
}

// Batch tracks the current local state of a
// batch and all its contained transactions.
// Batch provides methods to handle instantiation
// of new `transaction.Pending`, validation of
// transactions based on the current batch state
// (gas limit, account balances, account nonces)
// and local inclusion/state application of
// transactions to the batch.
type Batch struct {
	mu       sync.RWMutex
	previous *Batch
	Broker   *medley.Broker[StateChangeResult]

	instanceID    uint64
	epochID       epochid.EpochID
	L1BlockNumber uint64
	ChainID       *big.Int
	decryptionKey []byte

	gasPool    core.GasPool
	block      sequencer.Block
	chainState sequencer.State

	committedTxs *transaction.Queue
	txPool       *transaction.Queue

	DecryptionKey  chan []byte
	ConfirmedBatch chan epochid.EpochID
	Transaction    chan *transaction.Pending
	subscription   chan StateChangeResult
	stopSignal     chan struct{}
	stoppedResult  chan error
}

func (b *Batch) Index() uint64 {
	return b.epochID.Big().Uint64()
}

func (b *Batch) EpochID() epochid.EpochID {
	return b.epochID
}

func (b *Batch) Log() *zerolog.Logger {
	logger := log.With().
		Dict("batch",
			zerolog.Dict().
				Str("epochID", b.EpochID().String()),
		).Logger()
	return &logger
}

// validateTx checks that the transaction `p` fulfills all preliminary
// conditions to be included in the batch.
// A valid transaction:
//
//  1. has a monotonically increasing nonce for the sender's account at the latest chain-state,
//     also considering all previous locally included transactions in that batch
//  2. has enough balance at the senders account in order to pay the tansactions gas fees, also
//     considering all previous locally included transactions in that batch
func (b *Batch) validateTx(ctx context.Context, p *transaction.Pending) error {
	currentNonce, err := b.chainState.GetNonce(ctx, p.Sender)
	if err != nil {
		return err
	}
	if p.Tx.Nonce() != currentNonce {
		return errors.Errorf("nonce mismatch, want: %d,got: %d", currentNonce, p.Tx.Nonce())
	}
	if err := ValidateGasParams(p.Tx, b.block.BaseFee()); err != nil {
		return err
	}
	p.GasCost = CalculateGasCost(p.Tx, b.block.BaseFee())
	p.MinerFee = CalculatePriorityFee(p.Tx, b.block.BaseFee())
	balance, err := b.chainState.GetBalance(ctx, p.Sender)
	if err != nil {
		return err
	}
	if balance.Cmp(p.GasCost) < 0 {
		return errors.New("not enough funds to pay gas fee")
	}
	return nil
}

// ApplyTx will include the transaction `p` in the local batch-state and
// will modify the batches local state to include the nonce and balance changes.
// ApplyTx can fail when the transaction's inclusion would surpass the batches
// gas limit.
func (b *Batch) ApplyTx(ctx context.Context, p *transaction.Pending) error {
	err := b.validateTx(ctx, p)
	if err != nil {
		return errors.Wrap(err, "validation failed")
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	err = b.gasPool.SubGas(p.Tx.Gas())
	if err != nil {
		// gas limit reached
		return err
	}
	newTotalSize := b.committedTxs.TotalByteSize() + len(p.TxBytes)
	if newTotalSize > BatchSizeLimit {
		return errors.Errorf("size limit reached (%d + %d > %d)",
			b.committedTxs.TotalByteSize(), len(p.TxBytes), BatchSizeLimit)
	}

	err = b.chainState.SubBalance(ctx, p.Sender, p.GasCost)
	if err != nil {
		return err
	}
	// not really necessary, only to e.g. observe the total gained fee
	err = b.chainState.AddBalance(ctx, b.block.Coinbase(), p.MinerFee)
	if err != nil {
		return err
	}
	b.chainState.SetNonce(p.Sender, p.Tx.Nonce()+1)
	b.committedTxs.Enqueue(p)
	return nil
}

func (b *Batch) Hash() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.committedTxs.Hash()
}

func (b *Batch) Run(ctx context.Context, epochTick <-chan time.Time) error {
	var (
		transition State
		err        error
	)

	logHandler := func(handler string, transition State) {
		b.Log().Debug().
			Str("handler", handler).
			Str("state", transition.StateEnum().String()).
			Msg("received value, calling handler")
	}
	transition = &Initial{}
	ctxDone := ctx.Done()
	for {
		stateChange := transition.Process(b)
		if stateChange != nil {
			stateChange.Log().Debug().Msg("state-change")
			b.Broker.Publish <- *stateChange
			next := transition.Post(b)
			if next.StateEnum() != transition.StateEnum() {
				transition = next
				// we transitioned directly,
				// skip the select
				continue
			} else if next.StateEnum() == NoState {
				b.stoppedResult <- err
				// transitions have been stopped.
				// this is only the case for Stopped{}
				close(b.Broker.Publish)
				b.Log().Debug().
					Str("state", transition.StateEnum().String()).
					Msg("stopped running")
				return err
			} else {
				transition = next
			}
		}
		select {
		case stateChange, ok := <-b.subscription:
			if !ok {
				b.subscription = nil
				continue
			}
			logHandler("OnStateChangePrevious", transition)
			transition = transition.OnStateChangePrevious(b, stateChange)
		case tim, ok := <-epochTick:
			if !ok {
				// disable channel,
				// nil channel will block on send/receive
				epochTick = nil
				continue
			}
			logHandler("OnEpochTick", transition)
			transition = transition.OnEpochTick(b, tim)
		case epochID, ok := <-b.ConfirmedBatch:
			if !ok {
				// disable channel,
				// nil channel will block on send/receive
				b.ConfirmedBatch = nil
				continue
			}
			logHandler("OnBatchConfirmation", transition)
			transition = transition.OnBatchConfirmation(b, epochID)
		case key, ok := <-b.DecryptionKey:
			if !ok {
				// disable channel,
				// nil channel will block on send/receive
				b.DecryptionKey = nil
				continue
			}
			logHandler("OnDecryptionKey", transition)
			transition = transition.OnDecryptionKey(b, key)
		case tx, ok := <-b.Transaction:
			if !ok {
				// disable channel,
				// nil channel will block on send/receive
				b.Transaction = nil
				continue
			}
			logHandler("OnTransaction", transition)
			transition = transition.OnTransaction(b, tx)
		case _, ok := <-b.stopSignal:
			if !ok {
				// disable channel,
				// nil channel will block on send/receive
				logHandler("OnStop", transition)
				transition = transition.OnStop(b)
			}
		case <-ctxDone:
			ctxDone = nil
			err = ctx.Err()
			logHandler("OnStop", transition)
			transition = transition.OnStop(b)
		}
	}
}

func (b *Batch) Stop() chan error {
	close(b.stopSignal)
	return b.stoppedResult
}
