package transaction

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	txtypes "github.com/shutter-network/txtypes/types"
	"gotest.tools/assert"
)

func TestTotalByteSize(t *testing.T) {
	signingKey, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)

	chainID := big.NewInt(1)
	signer := txtypes.LatestSignerForChainID(chainID)

	q := NewQueue()
	assert.Equal(t, q.TotalByteSize(), 0)

	for _, payloadSize := range []int{0, 10, 100, 1000} {
		txData := &txtypes.ShutterTx{
			ChainID:          chainID,
			Nonce:            2,
			GasTipCap:        big.NewInt(3),
			GasFeeCap:        big.NewInt(4),
			Gas:              5,
			EncryptedPayload: bytes.Repeat([]byte("x"), payloadSize),
			BatchIndex:       6,
		}
		tx, err := txtypes.SignNewTx(signingKey, signer, txData)
		assert.NilError(t, err)
		txBytes, err := tx.MarshalBinary()
		assert.NilError(t, err)
		pendingTx, err := NewPending(signer, txBytes, time.Now())
		assert.NilError(t, err)

		sizeBefore := q.TotalByteSize()
		q.Enqueue(pendingTx)
		assert.Equal(t, q.TotalByteSize(), sizeBefore+len(txBytes))
	}
}
