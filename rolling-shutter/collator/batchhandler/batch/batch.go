package batch

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
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

func NewCachedPendingBatch(
	ctx context.Context, epochID epochid.EpochID, l1BlockNumber uint64, client *ethclient.Client,
) (*Batch, error) {
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	signer := txtypes.LatestSignerForChainID(chainID)

	// batchindex not necessarily the same as the l2blocknumber.
	// just query for the current state of the addresses balance/nonce.
	// since the collator is the only one progressing the balance/nonce state,
	// this is fine as long as the current batch is not submitted
	// number=nil means the latest state from the node
	block, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	state := sequencer.NewChainBatchCache(client, nil)
	b := &Batch{
		ChainID:       chainID,
		epochID:       epochID,
		l1BlockNumber: l1BlockNumber,
		signer:        signer,
		state:         state,
		block:         block,
		transactions:  transaction.NewTransactionQueue(),
	}
	b.gasPool.AddGas(block.GasLimit())
	return b, nil
}

// Batch tracks the current local state of a
// batch and all its contained transactions.
// Batch provides methods to handle instantiation
// of new `PendingTransaction`, validation of
// transactions based on the current batch state
// (gas limit, account balances, account nonces)
// and local inclusion/state application of
// transactions to the batch.
type Batch struct {
	mux sync.Mutex

	gasPool      core.GasPool
	block        sequencer.Block
	signer       txtypes.Signer
	state        sequencer.State
	transactions *transaction.TransactionQueue

	epochID       epochid.EpochID
	l1BlockNumber uint64
	ChainID       *big.Int
}

func (b *Batch) BatchIndex() (uint64, error) {
	i := b.epochID.Big()
	if !i.IsUint64() {
		return 0, errors.Errorf("epoch id %s does not represent a batch index", b.epochID)
	}
	return i.Uint64(), nil
}

func (b *Batch) EpochID() epochid.EpochID {
	return b.epochID
}

// ValidateTx checks that the transaction `p` fulfills all preliminary
// conditions to be included in the batch.
// A valid transaction:
//    a) has a monotonically increasing nonce for the sender's
//        account at the latest chain-state, also considering all
//        previous locally
//        included transactions in that batch
//    b) has enough balance at the senders account in order to pay the
//        tansactions gas fees, also considering all previous locally
//        included transactions in that batch
func (b *Batch) ValidateTx(ctx context.Context, p *transaction.Pending) error {
	currentNonce, err := b.state.GetNonce(ctx, p.Sender)
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
	balance, err := b.state.GetBalance(ctx, p.Sender)
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
	b.mux.Lock()
	defer b.mux.Unlock()

	err := b.gasPool.SubGas(p.Tx.Gas())
	if err != nil {
		// gas limit reached
		return err
	}
	newTotalSize := b.transactions.TotalByteSize() + len(p.TxBytes)
	if newTotalSize > BatchSizeLimit {
		return errors.Errorf("size limit reached (%d + %d > %d)",
			b.transactions.TotalByteSize(), len(p.TxBytes), BatchSizeLimit)
	}

	err = b.state.SubBalance(ctx, p.Sender, p.GasCost)
	if err != nil {
		return err
	}
	// not really necessary, only to e.g. observe the total gained fee
	err = b.state.AddBalance(ctx, b.block.Coinbase(), p.MinerFee)
	if err != nil {
		return err
	}
	b.state.SetNonce(p.Sender, p.Tx.Nonce()+1)

	b.transactions.Enqueue(p)
	return nil
}

func (b *Batch) Transactions() *transaction.TransactionQueue {
	return b.transactions
}

// SignedBatchTx constructs the signed Batch-Transaction to be sent to
// the sequencer for batch submittal.
func (b *Batch) SignedBatchTx(privateKey *ecdsa.PrivateKey, decryptionKey []byte) (*txtypes.Transaction, error) {
	txs := b.Transactions()
	ts := time.Now().Unix()
	batchIndex, err := b.BatchIndex()
	if err != nil {
		return nil, err
	}
	btxData := &txtypes.BatchTx{
		ChainID:       b.ChainID,
		DecryptionKey: decryptionKey,
		BatchIndex:    batchIndex,
		L1BlockNumber: b.l1BlockNumber,
		Timestamp:     big.NewInt(ts),
		Transactions:  txs.Bytes(),
	}
	return txtypes.SignNewTx(privateKey, b.signer, btxData)
}
