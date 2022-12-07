package batcher

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/shutter-network/txtypes/types"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/batchhandler/batch"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/epochid"
)

var (
	ErrNonceMismatch         = errors.New("nonce mismatch")
	ErrAccountNotInitialized = errors.New("account not initialized")
	ErrCannotPayGasFee       = errors.New("not enough funds to pay gas fee")
	ErrGasLimitReached       = errors.New("gas limit reached")
	ErrBatchSizeLimitReached = errors.New("batch size limit reached")
)

// ChainState is used by the collator to simulate applying transactions to a per block state and
// decide if transactions will be included into the next block. The collator creates a new
// ChainState object for each block and calls ApplyTx for each transaction that it wants to apply.
type ChainState struct {
	balances        map[common.Address]*big.Int
	nonces          map[common.Address]uint64
	gasUsed         uint64
	sizeInBytes     uint64
	numTransactions uint64
	signer          txtypes.Signer
	baseFee         *big.Int
	blockGasLimit   uint64
	epochID         epochid.EpochID
}

func NewChainState(signer txtypes.Signer, baseFee *big.Int, blockGasLimit uint64, epochID epochid.EpochID) *ChainState {
	return &ChainState{
		balances:      make(map[common.Address]*big.Int),
		nonces:        make(map[common.Address]uint64),
		gasUsed:       0,
		signer:        signer,
		baseFee:       baseFee,
		blockGasLimit: blockGasLimit,
		epochID:       epochID,
	}
}

// IsAccountInitialized returns true iff the given account has already been initialized.
func (chst *ChainState) IsAccountInitialized(account common.Address) bool {
	_, ok := chst.balances[account]
	return ok
}

// InitializeAccount initializes the given account with the given balance and nonce.
func (chst *ChainState) InitializeAccount(account common.Address, balance *big.Int, nonce uint64) {
	chst.balances[account] = balance
	chst.nonces[account] = nonce
}

// CanApplyTx checks if the transaction can be applied to the state. The caller must have already
// verified some basic properties (chainId matches, sender can be extracted) and the account must
// have already been initialized with InitializeAccount.
func (chst *ChainState) CanApplyTx(tx *txtypes.Transaction, txSizeInBytes uint64) error {
	account, err := chst.signer.Sender(tx)
	if err != nil {
		return err
	}
	if !chst.IsAccountInitialized(account) {
		return ErrAccountNotInitialized
	}

	if tx.Nonce() != chst.nonces[account] {
		return ErrNonceMismatch
	}

	err = batch.ValidateGasParams(tx, chst.baseFee)
	if err != nil {
		return err
	}

	gasCost := batch.CalculateGasCost(tx, chst.baseFee)
	if chst.balances[account].Cmp(gasCost) < 0 {
		return ErrCannotPayGasFee
	}

	if chst.gasUsed+tx.Gas() > chst.blockGasLimit {
		return ErrGasLimitReached
	}

	if chst.sizeInBytes+txSizeInBytes > batch.BatchSizeLimit {
		return ErrBatchSizeLimitReached
	}

	return nil
}

// ApplyTx applies the given transaction. The caller must have called CanApplyTx first.
func (chst *ChainState) ApplyTx(tx *txtypes.Transaction, txSizeInBytes uint64) {
	account, err := chst.signer.Sender(tx)
	if err != nil {
		panic(err)
	}

	chst.gasUsed += tx.Gas()
	chst.sizeInBytes += txSizeInBytes

	chst.nonces[account]++

	gasCost := batch.CalculateGasCost(tx, chst.baseFee)
	balance := chst.balances[account]
	balance.Sub(balance, gasCost)
	chst.numTransactions++
}
