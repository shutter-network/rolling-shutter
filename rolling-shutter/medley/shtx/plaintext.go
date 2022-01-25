package shtx

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

type PlaintextTransaction struct {
	Receiver           common.Address
	Calldata           []byte
	Value              *big.Int
	GasLimit           *big.Int
	InclusionFeePerGas *big.Int
	ExecutionFeePerGas *big.Int
	Nonce              *big.Int
	Signature          []byte
}

func decodePlaintextTransactionPayload(payload []byte) (*CipherTransaction, error) {
	return nil, nil
}

func (t *PlaintextTransaction) Type() uint8 {
	return PlaintextTransactionType
}

func (t *PlaintextTransaction) Encode() ([]byte, error) {
	// PlaintextTransaction consists only of byte slices and big ints, both of which are supported
	// by rlp out of the box. An error is returned if one of the integers is negative
	return rlp.EncodeToBytes(t)
}

func (t *PlaintextTransaction) EnvelopeSigner() (common.Address, error) {
	return common.Address{}, nil
}
