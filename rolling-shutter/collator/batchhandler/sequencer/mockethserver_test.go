package sequencer

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
)

func newTestConfig(t *testing.T) config.Config {
	t.Helper()

	ethereumKey, err := ethcrypto.GenerateKey()
	assert.NilError(t, err)
	return config.Config{
		EthereumURL:         "http://127.0.0.1:8454",
		SequencerURL:        "http://127.0.0.1:8455",
		EthereumKey:         ethereumKey,
		ExecutionBlockDelay: uint32(5),
		InstanceID:          123,
	}
}

func TestMockEthIntegration(t *testing.T) {
	cfg := newTestConfig(t)
	eth := RunMockEthServer(t)
	defer eth.Teardown()

	rpcClient, err := rpc.Dial(eth.URL)
	assert.NilError(t, err)
	ctx := context.Background()
	ethClient := ethclient.NewClient(rpcClient)

	address := ethcrypto.PubkeyToAddress(cfg.EthereumKey.PublicKey)
	balance := big.NewInt(42)
	nonce := uint64(42)
	chainID := big.NewInt(0)
	baseFee := big.NewInt(1000000)
	gasLimit := uint64(10000000)
	coinbase := common.HexToAddress("0x0000000000000000000000000000000000000000")

	// Set the values on the dummy rpc server
	eth.SetBalance(address, balance, "latest")
	eth.SetNonce(address, nonce, "latest")
	eth.SetChainID(chainID)
	eth.SetBlock(baseFee, gasLimit, "latest")

	// Use the client lib to poll the dummy rpc server
	polledBalance, err := ethClient.BalanceAt(ctx, address, nil)
	assert.NilError(t, err)

	polledNonce, err := ethClient.NonceAt(ctx, address, nil)
	assert.NilError(t, err)

	polledChainID, err := ethClient.ChainID(ctx)
	assert.NilError(t, err)

	polledBlock, err := ethClient.BlockByNumber(ctx, nil)
	assert.NilError(t, err)

	// Assert equal values to what was provided to the dummy rpc server
	assert.Assert(t, balance.Cmp(polledBalance) == 0)

	assert.Equal(t, gasLimit, polledBlock.GasLimit())
	assert.Assert(t, baseFee.Cmp(polledBlock.BaseFee()) == 0)
	assert.Equal(t, polledBlock.Coinbase(), coinbase)

	assert.Equal(t, polledNonce, nonce)

	assert.Assert(t, chainID.Cmp(polledChainID) == 0)
}
