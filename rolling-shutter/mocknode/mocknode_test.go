package mocknode

import (
	"math/big"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	gocmp "github.com/google/go-cmp/cmp"
	txtypes "github.com/shutter-network/txtypes/types"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

func TestSaltedEncryption(t *testing.T) {
	p, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	address := ethcrypto.PubkeyToAddress(p.PublicKey)

	payload := &txtypes.ShutterPayload{
		To:    &address,
		Value: big.NewInt(42),
	}
	payloadBytes, err := payload.Encode()
	assert.NilError(t, err)

	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(0)

	_, eonPublicKey, err := computeEonKeys(42)
	assert.NilError(t, err)

	enc1, err := EncryptMessage(payloadBytes, identityPreimage, eonPublicKey)
	assert.NilError(t, err)
	enc2, err := EncryptMessage(payloadBytes, identityPreimage, eonPublicKey)
	assert.NilError(t, err)
	assert.Assert(t, !gocmp.Equal(enc1, enc2))
}
