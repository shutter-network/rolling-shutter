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

func DecodePlaintextTx(input []byte) (*PlaintextTransaction, error) {
	tx := &PlaintextTransaction{}
	err := rlp.DecodeBytes(input[1:], tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (t *PlaintextTransaction) Type() uint8 {
	return PlaintextTransactionType
}

func (t *PlaintextTransaction) Encode() ([]byte, error) {
	// PlaintextTransaction consists only of byte slices and big ints, both of which are supported
	// by rlp out of the box. An error is returned if one of the integers is negative
	rlpEncoded, err := rlp.EncodeToBytes(t)
	if err != nil {
		return nil, err
	}
	return append([]byte{PlaintextTransactionType}, rlpEncoded...), nil
}

func (t *PlaintextTransaction) EnvelopeSigner() (common.Address, error) {
	return common.Address{}, nil
}
