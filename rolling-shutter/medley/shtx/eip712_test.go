package shtx

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core"
	"gotest.tools/assert"
)

func TestEncodeData(t *testing.T) {
	// values taken from EIP712 example script
	// see: https://github.com/ethereum/EIPs/blob/a2b62466f7b02edcdae5725dba028f4935f41ed5/assets/eip-712/Example.js#L116
	testTypedData := core.TypedData{
		Types: core.Types{
			EIP712Domain: EIP712DomainType,
			"Person": {
				{Name: "name", Type: "string"},
				{Name: "wallet", Type: "address"},
			},
			"Mail": {
				{Name: "from", Type: "Person"},
				{Name: "to", Type: "Person"},
				{Name: "contents", Type: "string"},
			},
		},
		PrimaryType: "Mail",
		Domain: core.TypedDataDomain{
			Name:              "Ether Mail",
			Version:           "1",
			ChainId:           math.NewHexOrDecimal256(1),
			VerifyingContract: "0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC",
		},
		Message: core.TypedDataMessage{
			"from":     map[string]interface{}{"name": "Cow", "wallet": "0xCD2a3d9F938E13CD947Ec05AbC7FE734Df8DD826"},
			"to":       map[string]interface{}{"name": "Bob", "wallet": "0xbBbBBBBbbBBBbbbBbbBbbbbBBbBbbbbBbBbbBBbB"},
			"contents": "Hello, Bob!",
		},
	}
	expectedDomainSeparator := "f2cee375fa42b42143804025fc449deafd50cc031ca257e0b194a650a912090f"
	expectedHashStruct := "c52c0ee5d84264471806290a3f2c4cecfc5490626bf912d01f240d7a274b371e"
	expectedEncoded, err := hex.DecodeString("1901" + expectedDomainSeparator + expectedHashStruct)
	assert.NilError(t, err)

	encoded, err := EIP712Encode(&testTypedData)
	assert.NilError(t, err)
	assert.Check(t, bytes.Equal(encoded, expectedEncoded))

	expectedHashToSign := "0xbe609aee343fb3c4b28e1df9e632fca64fcfaede20f02e86244efddf30957bd2"
	hashToSign, err := HashForSigning(&testTypedData)
	assert.NilError(t, err)
	assert.Equal(t, hashToSign.String(), expectedHashToSign)
}
