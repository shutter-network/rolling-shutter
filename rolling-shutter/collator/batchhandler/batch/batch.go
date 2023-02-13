package batch

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"
)

const (
	minimumTxGas   uint64 = 21000
	BatchSizeLimit        = 8 * 1024
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
