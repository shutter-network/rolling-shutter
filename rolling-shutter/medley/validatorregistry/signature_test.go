package validatorregistry

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	blst "github.com/supranational/blst/bindings/go"
)

func TestSignature(t *testing.T) {
	msg := &RegistrationMessage{
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
