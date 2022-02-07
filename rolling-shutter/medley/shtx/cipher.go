package shtx

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/signer/core"
)

type CipherTransaction struct {
	EncryptedPayload   []byte
	GasLimit           *big.Int
	InclusionFeePerGas *big.Int
	ExecutionFeePerGas *big.Int
	Nonce              *big.Int
	Signature          []byte
}

var eip712CipherTransactionType = []core.Type{
	{Name: "EncryptedPayload", Type: "bytes"},
	{Name: "GasLimit", Type: "uint256"},
	{Name: "InclusionFeePerGas", Type: "uint256"},
	{Name: "ExecutionFeePerGas", Type: "uint256"},
	{Name: "Nonce", Type: "uint256"},
}

var eip712ShutterDomain = core.TypedDataDomain{
	Name:    "shutter",
	Version: "1",
}

func decodeCipherTransaction(input []byte) (*CipherTransaction, error) {
	tx := &CipherTransaction{}
	// We cut out the first byte that represents the transaction type
	err := rlp.DecodeBytes(input[1:], tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (t *CipherTransaction) Type() uint8 {
	return CipherTransactionType
}

func (t *CipherTransaction) Encode() ([]byte, error) {
	// CipherTransaction consists only of byte slices and big ints, both of which are supported
	// by rlp out of the box. An error is returned if one of the integers is negative
	rlpEncoded, err := rlp.EncodeToBytes(t)
	if err != nil {
		return nil, err
	}
	return append([]byte{CipherTransactionType}, rlpEncoded...), nil
}

func (t *CipherTransaction) EnvelopeSigner() (common.Address, error) {
	hash, err := t.SigningHash()
	if err != nil {
		return common.Address{}, err
	}
	publicKey, err := crypto.SigToPub(hash.Bytes(), t.Signature)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*publicKey), nil
}

func (t *CipherTransaction) SigningHash() (common.Hash, error) {
	typedDataTransaction := core.TypedData{
		Types: core.Types{
			EIP712Domain:        ShortEIP712DomainType,
			"CipherTransaction": eip712CipherTransactionType,
		},
		PrimaryType: "CipherTransaction",
		Domain:      eip712ShutterDomain,
		Message: core.TypedDataMessage{
			"EncryptedPayload":   t.EncryptedPayload,
			"GasLimit":           (*math.HexOrDecimal256)(t.GasLimit),
			"InclusionFeePerGas": (*math.HexOrDecimal256)(t.InclusionFeePerGas),
			"ExecutionFeePerGas": (*math.HexOrDecimal256)(t.ExecutionFeePerGas),
			"Nonce":              (*math.HexOrDecimal256)(t.Nonce),
		},
	}
	return HashForSigning(&typedDataTransaction)
}
