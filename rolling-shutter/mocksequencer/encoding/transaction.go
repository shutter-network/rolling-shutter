package encoding

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

// txJSON is the JSON representation of transactions.
type TransactionJSON struct {
	Type hexutil.Uint64 `json:"type"`

	// Common transaction fields:
	Nonce                hexutil.Uint64 `json:"nonce"`
	GasPrice             *hexutil.Big   `json:"gasPrice"`
	MaxPriorityFeePerGas *hexutil.Big   `json:"maxPriorityFeePerGas"`
	MaxFeePerGas         *hexutil.Big   `json:"maxFeePerGas"`
	Gas                  hexutil.Uint64 `json:"gas"`
	Value                *hexutil.Big   `json:"value"`
	// XXX for some reason hardhat's eth_sendTransaction
	// expects a "data" field instead of "input"
	// (as per the JSON RPC spec)
	// Data                 hexutil.Bytes   `json:"input"`
	Data hexutil.Bytes   `json:"data"`
	To   *common.Address `json:"to"`
	From *common.Address `json:"from"`

	// Access list transaction fields:
	ChainID    *hexutil.Big      `json:"chainId,omitempty"`
	AccessList *types.AccessList `json:"accessList,omitempty"`
}

// MarshalJSON marshals as JSON with a hash.
func ToTransactionJSON(tx *types.Transaction, from *common.Address) *TransactionJSON {
	enc := &TransactionJSON{}
	// These are set for all tx types.
	// enc.Hash = tx.Hash()
	enc.Type = hexutil.Uint64(tx.Type())

	enc.ChainID = (*hexutil.Big)(tx.ChainId())
	al := tx.AccessList()
	enc.AccessList = &al
	enc.Nonce = hexutil.Uint64(tx.Nonce())
	enc.Gas = hexutil.Uint64(tx.Gas())
	enc.MaxFeePerGas = (*hexutil.Big)(tx.GasFeeCap())
	enc.MaxPriorityFeePerGas = (*hexutil.Big)(tx.GasTipCap())
	enc.Value = (*hexutil.Big)(tx.Value())
	enc.Data = hexutil.Bytes(tx.Data())
	enc.To = tx.To()
	enc.From = from
	return enc
}
