package batch

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gotest.tools/assert"
)

func TestCaching(t *testing.T) {
	eth := RunMockEthServer(t)
	defer eth.Teardown()

	client, err := ethclient.Dial(eth.URL)
	assert.NilError(t, err)

	var b []byte
	addr := common.BytesToAddress(b)
	bal1 := big.NewInt(42000)
	nonce1 := uint64(42)

	eth.SetBalance(addr, bal1, "latest")
	eth.SetNonce(addr, nonce1, "latest")

	ctx := context.Background()
	cbc := &ChainBatchCache{
		balances:      make(map[common.Address]*big.Int, 0),
		nonces:        make(map[common.Address]uint64, 0),
		Client:        client,
		AtBlockNumber: nil, // nil means latest block
	}

	// initial state should be polled from the client
	bal, err := cbc.GetBalance(ctx, addr)
	assert.NilError(t, err)
	assert.Check(
		t,
		bal1.Cmp(bal) == 0)

	nonce, err := cbc.GetNonce(ctx, addr)
	assert.NilError(t, err)
	assert.Equal(
		t,
		nonce,
		nonce1,
	)

	nonce2 := nonce1 + 1
	diff := big.NewInt(2000)

	cbc.SetNonce(addr, nonce2)
	err = cbc.SubBalance(ctx, addr, diff)
	assert.NilError(t, err)
	// nulls out, but we call add as well
	err = cbc.SubBalance(ctx, addr, diff)
	assert.NilError(t, err)
	// nulls out, but we call add as well
	err = cbc.AddBalance(ctx, addr, diff)
	assert.NilError(t, err)
	// nulls out, but we call add as well

	// Client nonce should stay the same
	// nulls out, but we call add as well
	polledNonce, err := client.NonceAt(ctx, addr, nil)
	assert.NilError(t, err)
	assert.Assert(
		t,
		polledNonce == nonce1,
	)
	// Client balance should stay the same
	polledBalance, err := client.BalanceAt(ctx, addr, nil)
	assert.NilError(t, err)
	assert.Assert(
		t,
		polledBalance.Cmp(bal1) == 0,
	)

	nonce, err = cbc.GetNonce(ctx, addr)
	assert.NilError(t, err)
	// Cache nonce should be updated
	assert.Equal(
		t,
		nonce,
		nonce2,
	)
	bal2 := big.NewInt(0).Sub(bal1, diff)
	bal, err = cbc.GetBalance(ctx, addr)
	assert.NilError(t, err)
	// Cache balance should be updated
	assert.Assert(
		t,
		bal.Cmp(bal2) == 0,
	)
}
