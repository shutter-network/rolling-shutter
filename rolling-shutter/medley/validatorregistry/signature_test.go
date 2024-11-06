package validatorregistry

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	blst "github.com/supranational/blst/bindings/go"
)

func TestSignature(t *testing.T) {
	msg := &LegacyRegistrationMessage{
		Version:                  1,
		ChainID:                  2,
		ValidatorRegistryAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ValidatorIndex:           3,
		Nonce:                    4,
		IsRegistration:           true,
	}

	var ikm [32]byte
	privkey := blst.KeyGen(ikm[:])
	pubkey := new(blst.P1Affine).From(privkey)

	sig := CreateSignature(privkey, msg)
	check := VerifySignature(sig, pubkey, msg)
	assert.True(t, check)

	msg.IsRegistration = false
	check = VerifySignature(sig, pubkey, msg)
	assert.False(t, check)
}

func TestAggSignature(t *testing.T) {
	msg := &AggregateRegistrationMessage{
		Version:                  1,
		ChainID:                  2,
		ValidatorRegistryAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ValidatorIndex:           3,
		Nonce:                    4,
		Count:                    2,
		IsRegistration:           true,
	}

	var ikm [32]byte
	var sks []*blst.SecretKey
	var pks []*blst.P1Affine
	for i := 0; i < int(msg.Count); i++ {
		privkey := blst.KeyGen(ikm[:])
		pubkey := new(blst.P1Affine).From(privkey)
		sks = append(sks, privkey)
		pks = append(pks, pubkey)
	}

	sig := CreateAggregateSignature(sks, msg)
	check := VerifyAggregateSignature(sig, pks, msg)
	assert.True(t, check)

	msg.IsRegistration = false
	check = VerifyAggregateSignature(sig, pks, msg)
	assert.False(t, check)
}
