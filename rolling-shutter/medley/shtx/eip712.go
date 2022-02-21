package shtx

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

var EIP712DomainType = []apitypes.Type{
	{Name: "name", Type: "string"},
	{Name: "version", Type: "string"},
	{Name: "chainId", Type: "uint256"},
	{Name: "verifyingContract", Type: "address"},
}

var ShortEIP712DomainType = []apitypes.Type{
	{Name: "name", Type: "string"},
	{Name: "version", Type: "string"},
}

const EIP712Domain = "EIP712Domain"

func EIP712Encode(typedData *apitypes.TypedData) ([]byte, error) {
	domainSeparator, err := typedData.HashStruct(EIP712Domain, typedData.Domain.Map())
	if err != nil {
		return nil, err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}
	rawData := bytes.Join(
		[][]byte{
			{0x19, 0x01},
			[]byte(domainSeparator),
			typedDataHash,
		},
		nil)
	return rawData, nil
}

func HashForSigning(typedData *apitypes.TypedData) (common.Hash, error) {
	encodedData, err := EIP712Encode(typedData)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(encodedData), nil
}
