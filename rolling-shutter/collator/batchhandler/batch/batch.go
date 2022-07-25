package batch

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/sequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/transaction"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
)

const (
	minimumTxGas   uint64 = 21000
	BatchSizeLimit int    = 8 * 1024
)

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
		Broker:         medley.NewBroker[StateChangeResult](1, true),
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
//    a) has a monotonically increasing nonce for the sender's
//        account at the latest chain-state, also considering all
//        previous locally
//        included transactions in that batch
//    b) has enough balance at the senders account in order to pay the
//        tansactions gas fees, also considering all previous locally
//        included transactions in that batch
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
