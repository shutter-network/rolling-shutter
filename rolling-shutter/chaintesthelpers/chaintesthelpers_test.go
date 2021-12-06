package chaintesthelpers

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"gotest.tools/assert"
)

func TestClientReverts(t *testing.T) {
	SkipChainTests(t)
	ctx := context.Background()
	client := GetChainClient(ctx, t)

	snapshotID := client.TakeSnapshot(ctx, t)
	snapshotBlk := client.GetBlockNumber(ctx, t)

	client.MineBlock(ctx, t)
	newBlk := client.GetBlockNumber(ctx, t)

	client.RevertToSnapshot(ctx, t, snapshotID)
	revertBlk := client.GetBlockNumber(ctx, t)

	assert.Equal(t, snapshotBlk+1, newBlk)
	assert.Equal(t, snapshotBlk, revertBlk)
}

func TestChainCleanup(t *testing.T) {
	SkipChainTests(t)
	ctx := context.Background()
	contracts, cleanup := NewTestContracts(ctx, t)
	client := contracts.Client

	snapshotBlk, err := client.BlockNumber(ctx)
	assert.NilError(t, err)

	// Get a signer and send a valid transaction to force minting a block
	privateKey, err := crypto.HexToECDSA(HardhatFundedKey)
	assert.NilError(t, err)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	assert.Check(t, ok)

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	assert.NilError(t, err)

	value := big.NewInt(1000000000000000000) // in wei (1 eth)
	gasLimit := uint64(21000)                // in units
	gasPrice, err := client.SuggestGasPrice(ctx)
	assert.NilError(t, err)

	tx := types.NewTransaction(nonce, fromAddress, value, gasLimit, gasPrice, []byte{})

	chainID, err := client.NetworkID(context.Background())
	assert.NilError(t, err)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	assert.NilError(t, err)

	err = client.SendTransaction(context.Background(), signedTx)
	assert.NilError(t, err)

	newBlk, err := client.BlockNumber(ctx)
	assert.NilError(t, err)

	cleanup()

	revertBlk, err := client.BlockNumber(ctx)
	assert.NilError(t, err)

	assert.Equal(t, snapshotBlk+1, newBlk)
	assert.Equal(t, snapshotBlk, revertBlk)
}
