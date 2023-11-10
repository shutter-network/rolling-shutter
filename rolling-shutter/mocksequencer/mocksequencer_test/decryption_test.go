package sequencer_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	txtypes "github.com/shutter-network/txtypes/types"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testlog"
)

func init() {
	testlog.Setup()
}

func TestServerDecryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	fixtures, err := NewFixtures(ctx, 1, true)
	assert.NilError(t, err)

	seqClient := fixtures.SequencerClient
	err = seqClient.SetBalance(ctx, fixtures.AddressSenders[0], big.NewInt(100000000000))
	assert.NilError(t, err)

	currentBatchIndex, err := seqClient.BatchIndex(ctx)
	assert.NilError(t, err)

	nextBatchIndex := currentBatchIndex + 1
	identityPreimage := identitypreimage.Uint64ToIdentityPreimage(nextBatchIndex)
	epochSecretKey, err := fixtures.KeyEnvironment.EpochSecretKey(identityPreimage.Bytes())
	assert.NilError(t, err)
	l1BlockNumber := uint64(42)

	fixtures.L1Service.setBlockNumber(l1BlockNumber)
	time.Sleep(sequencerL1PollInterval + 200*time.Millisecond)
	// HACK this is not the best way to synchronize.
	// Better would be to keep a last-time-polled counter in the sequencer
	// and here just wait for the next one

	shutterTxInner, err := fixtures.MakeShutterTx(0, nextBatchIndex, l1BlockNumber, nil)
	assert.NilError(t, err)

	shutterTx, err := txtypes.SignNewTx(fixtures.PrivkeySenders[0], fixtures.Signer, shutterTxInner)
	assert.NilError(t, err)
	shtxBytes, err := shutterTx.MarshalBinary()
	assert.NilError(t, err)

	batchTxInner := txtypes.BatchTx{
		ChainID:       fixtures.ChainID,
		DecryptionKey: epochSecretKey.Marshal(),
		BatchIndex:    nextBatchIndex,
		L1BlockNumber: l1BlockNumber,
		Timestamp:     big.NewInt(time.Now().Unix()),
		Transactions:  [][]byte{shtxBytes},
	}
	batchTx, err := txtypes.SignNewTx(fixtures.PrivkeyCollator, fixtures.Signer, &batchTxInner)
	assert.NilError(t, err)

	txHash, err := seqClient.SubmitBatch(ctx, batchTx)
	assert.NilError(t, err)
	assert.DeepEqual(t, txHash.Hex(), batchTx.Hash().Hex())

	finalisedShutterTx, isPending, err := seqClient.TransactionByHash(ctx, shutterTx.Hash())
	assert.NilError(t, err)

	assert.Equal(t, finalisedShutterTx.Type(), uint8(txtypes.ShutterTxType))
	assert.Equal(t, shutterTx.Hash().Hex(), finalisedShutterTx.Hash().Hex())
	assert.Equal(t, isPending, false)

	// compare equality of the underlying data structs
	// (this includes the decrypted payload)
	finalisedShutterTxInner, ok := finalisedShutterTx.TxInner().(*txtypes.ShutterTx)
	assert.Assert(t, ok)

	// exclude the signature, since the initial ShutterTx inner was not signed
	assert.DeepEqual(t,
		shutterTxInner,
		finalisedShutterTxInner,
		BigIntComparer,
		cmpopts.IgnoreFields(txtypes.ShutterTx{},
			"R",
			"V",
			"S",
		),
	)

	// make sure there is a decrypted payload and the outer transaction's
	// getters have access to the values
	assert.DeepEqual(t, finalisedShutterTx.To(), shutterTxInner.Payload.To)
	assert.DeepEqual(t, finalisedShutterTx.Value(), shutterTxInner.Payload.Value, BigIntComparer)
	assert.DeepEqual(t, finalisedShutterTx.Data(), shutterTxInner.Payload.Data)

	// Recover the signature
	recoveredSender, err := fixtures.Signer.Sender(finalisedShutterTx)
	assert.NilError(t, err)
	assert.DeepEqual(t, fixtures.AddressSenders[0], recoveredSender)
}
