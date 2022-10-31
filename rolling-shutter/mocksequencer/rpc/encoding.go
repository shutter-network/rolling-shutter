package rpc

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	txtypes "github.com/shutter-network/txtypes/types"
)

type RPCTransaction struct {
	BlockHash        *common.Hash        `json:"blockHash"`
	BlockNumber      *hexutil.Big        `json:"blockNumber"`
	From             common.Address      `json:"from"`
	Gas              hexutil.Uint64      `json:"gas"`
	GasPrice         *hexutil.Big        `json:"gasPrice"`
	GasFeeCap        *hexutil.Big        `json:"maxFeePerGas,omitempty"`
	GasTipCap        *hexutil.Big        `json:"maxPriorityFeePerGas,omitempty"`
	Hash             common.Hash         `json:"hash"`
	Input            hexutil.Bytes       `json:"input"`
	Nonce            hexutil.Uint64      `json:"nonce"`
	To               *common.Address     `json:"to"`
	TransactionIndex *hexutil.Uint64     `json:"transactionIndex"`
	Value            *hexutil.Big        `json:"value"`
	Type             hexutil.Uint64      `json:"type"`
	Accesses         *txtypes.AccessList `json:"accessList,omitempty"`
	ChainID          *hexutil.Big        `json:"chainId,omitempty"`
	V                *hexutil.Big        `json:"v"`
	R                *hexutil.Big        `json:"r"`
	S                *hexutil.Big        `json:"s"`
}

func newRPCTransaction(tx *txtypes.Transaction, blockHash common.Hash, blockNumber uint64, index uint64, baseFee *big.Int, signer txtypes.Signer) *RPCTransaction {
	// Copied and modified from github.com/ethereum/go-ethereum.

	from, _ := txtypes.Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()

	result := &RPCTransaction{
		Type:     hexutil.Uint64(tx.Type()),
		From:     from,
		Gas:      hexutil.Uint64(tx.Gas()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data()),
		Nonce:    hexutil.Uint64(tx.Nonce()),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Value()),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = &blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = (*hexutil.Uint64)(&index)
	}
	switch tx.Type() {
	case txtypes.LegacyTxType:
		// if a legacy transaction has an EIP-155 chain id, include it explicitly
		if id := tx.ChainId(); id.Sign() != 0 {
			result.ChainID = (*hexutil.Big)(id)
		}
	case txtypes.AccessListTxType:
		al := tx.AccessList()
		result.Accesses = &al
		result.ChainID = (*hexutil.Big)(tx.ChainId())
	case txtypes.DynamicFeeTxType:
		al := tx.AccessList()
		result.Accesses = &al
		result.ChainID = (*hexutil.Big)(tx.ChainId())
		result.GasFeeCap = (*hexutil.Big)(tx.GasFeeCap())
		result.GasTipCap = (*hexutil.Big)(tx.GasTipCap())
		// if the transaction has been mined, compute the effective gas price
		if baseFee != nil && blockHash != (common.Hash{}) {
			// FIXME this might be different for rolling-shutter
			// price = min(tip, gasFeeCap - baseFee) + baseFee
			price := math.BigMin(new(big.Int).Add(tx.GasTipCap(), baseFee), tx.GasFeeCap())
			result.GasPrice = (*hexutil.Big)(price)
		} else {
			result.GasPrice = (*hexutil.Big)(tx.GasFeeCap())
		}
	}
	return result
}

func stringToAddress(addr string) (common.Address, error) {
	if !common.IsHexAddress(addr) {
		var a common.Address
		return a, errors.New("address is no hex address string, can't decode")
	}
	return common.HexToAddress(addr), nil
}
