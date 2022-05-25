package batch

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	txtypes "github.com/shutter-network/txtypes/types"
)

type Block interface {
	BaseFee() *big.Int
	Coinbase() common.Address
	Number() *big.Int
}

type State interface {
	GetBalance(ctx context.Context, a common.Address) (*big.Int, error)
	SubBalance(ctx context.Context, a common.Address, diff *big.Int) error
	AddBalance(ctx context.Context, a common.Address, diff *big.Int) error
	GetNonce(ctx context.Context, a common.Address) (uint64, error)
	SetNonce(a common.Address, nonce uint64)
}

type EthClient interface {
	BalanceAt(context.Context, common.Address, *big.Int) (*big.Int, error)
	NonceAt(context.Context, common.Address, *big.Int) (uint64, error)
}

// ChainBatchCache tracks the state of account's nonces and balances
// for a certain Batch.
// If an address is not cached yet, it polls initial balances or nonces
// on a GetBalance or GetNonce call from the underlying ethereum node via the
// ChainBatchCache.Client. Then the value is cached and never polled again for
// that address.
// This allows to poll chain-state and then modify it locally, e.g. while
// accepting user transactions to be proposed as the next block to the sequencer.
type ChainBatchCache struct {
	balances map[common.Address]*big.Int
	nonces   map[common.Address]uint64

	Client        EthClient
	AtBlockNumber *big.Int
}

// GetBalance polls and caches the state of account `a` balance at the
// block number ChainBatchCache.AtBlockNumber.
func (c *ChainBatchCache) GetBalance(ctx context.Context, a common.Address) (*big.Int, error) {
	var err error

	bal, exists := c.balances[a]
	if !exists {
		bal, err = c.Client.BalanceAt(ctx, a, c.AtBlockNumber)
		if err != nil {
			return nil, err
		}
		c.balances[a] = bal
	}
	return bal, nil
}

// SubBalance subtracts the value `diff` from the balance of account `a`.
// If no balance is cached yet, SubBalance will conduct a call to the ethereum
// node to get the state of the balance before modifying it.
// The modified value is then persisted in the internal state cache.
func (c *ChainBatchCache) SubBalance(ctx context.Context, a common.Address, diff *big.Int) error {
	old, err := c.GetBalance(ctx, a)
	if err != nil {
		return err
	}
	newBal := new(big.Int).Sub(old, diff)
	if newBal.Sign() == -1 {
		return errors.New("subtracted balance would be negative")
	}
	c.balances[a] = newBal
	return nil
}

// AddBalance adds the value `diff` to the balance of account `a`.
// If no balance is cached yet, AddBalance will conduct a call to the ethereum
// node to get the state of the balance before modifying it.
// The modified value is then persisted in the internal state cache.
func (c *ChainBatchCache) AddBalance(ctx context.Context, a common.Address, diff *big.Int) error {
	old, err := c.GetBalance(ctx, a)
	if err != nil {
		return err
	}
	c.balances[a] = new(big.Int).Add(old, diff)
	return nil
}

// GetNonce polls and caches the state of account `a` balance at the
// block number ChainBatchCache.AtBlockNumber.
func (c *ChainBatchCache) GetNonce(ctx context.Context, a common.Address) (uint64, error) {
	var err error
	nonce, exists := c.nonces[a]
	if !exists {
		nonce, err = c.Client.NonceAt(ctx, a, c.AtBlockNumber)
		if err != nil {
			return nonce, err
		}
		c.nonces[a] = nonce
	}
	return nonce, nil
}

// SetNonce sets the value in the nonce cache of account `a` to value `nonce`.
// Once this is set, a call to GetNonce() will not poll the node but simply
// return the set value.
func (c *ChainBatchCache) SetNonce(a common.Address, nonce uint64) {
	c.nonces[a] = nonce
}

func (c *ChainBatchCache) Purge(a common.Address) {
	delete(c.nonces, a)
	delete(c.balances, a)
}

var _ State = (*ChainBatchCache)(nil)
