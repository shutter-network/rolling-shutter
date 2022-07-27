package sequencer

import (
	"context"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
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

type countedMux struct {
	mux *sync.RWMutex
	// the same counter is used for read/write
	// lock access, since the purpose of the counter
	// is simply to purge anused mutex objects from
	// the map to avoid unneccesary memory growth
	cnt uint32
}

type KeyedMutex[T comparable] struct {
	mu   sync.RWMutex
	muxs map[T]*countedMux
}

func NewKeyedMutex[T comparable]() *KeyedMutex[T] {
	return &KeyedMutex[T]{mu: sync.RWMutex{}, muxs: make(map[T]*countedMux)}
}

func (ml *KeyedMutex[T]) getMux(v T, releaseOperation bool) *sync.RWMutex {
	delta := 1
	if releaseOperation {
		delta = -1
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()

	vmu, ok := ml.muxs[v]
	if !ok {
		vmu = &countedMux{
			mux: &sync.RWMutex{},
			cnt: 0,
		}
		ml.muxs[v] = vmu
	}
	var addUint32 uint32
	if delta < 0 {
		// recommended operation to substract from counter
		addUint32 = ^uint32((-delta) - 1)
	} else {
		addUint32 = uint32(delta)
	}
	atomic.AddUint32(&vmu.cnt, addUint32)
	return vmu.mux
}

func (ml *KeyedMutex[T]) purgeDangling(v T) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	vmu, ok := ml.muxs[v]
	if !ok {
		return
	}
	if vmu.cnt <= 0 {
		delete(ml.muxs, v)
	}
}

func (ml *KeyedMutex[T]) Lock(v T) {
	mux := ml.getMux(v, false)
	mux.Lock()
}

func (ml *KeyedMutex[T]) Unlock(v T) {
	mux := ml.getMux(v, true)
	mux.Unlock()
	ml.purgeDangling(v)
}

func (ml *KeyedMutex[T]) RLock(v T) {
	mux := ml.getMux(v, false)
	mux.RLock()
}

func (ml *KeyedMutex[T]) RUnlock(v T) {
	mux := ml.getMux(v, true)
	mux.RUnlock()
	ml.purgeDangling(v)
}

func NewCached(client EthClient, atBlockNumber *big.Int) *Cached {
	return &Cached{
		balances:       make(map[common.Address]*big.Int),
		balancesLocks:  NewKeyedMutex[common.Address](),
		balancesMapMux: sync.RWMutex{},
		nonces:         make(map[common.Address]uint64),
		noncesLocks:    NewKeyedMutex[common.Address](),
		noncesMapMux:   sync.RWMutex{},
		Client:         client,
		AtBlockNumber:  atBlockNumber,
	}
}

// Cached tracks the state of account's nonces and balances
// for a certain Batch.
// If an address is not cached yet, it polls initial balances or nonces
// on a GetBalance or GetNonce call from the underlying ethereum node via the
// Cached.Client. Then the value is cached and never polled again for
// that address.
// This allows to poll chain-state and then modify it locally, e.g. while
// accepting user transactions to be proposed as the next block to the sequencer.
type Cached struct {
	balances       map[common.Address]*big.Int
	balancesLocks  *KeyedMutex[common.Address]
	balancesMapMux sync.RWMutex
	nonces         map[common.Address]uint64
	noncesLocks    *KeyedMutex[common.Address]
	noncesMapMux   sync.RWMutex

	Client        EthClient
	AtBlockNumber *big.Int
}

// GetBalance polls and caches the state of account `a` balance at the
// block number ChainBatchCache.AtBlockNumber.
func (c *Cached) GetBalance(ctx context.Context, a common.Address) (*big.Int, error) {
	// write lock because we eventually write to the dict
	c.balancesLocks.Lock(a)
	defer c.balancesLocks.Unlock(a)
	return c.getBalance(ctx, a)
}

func (c *Cached) getBalance(ctx context.Context, a common.Address) (*big.Int, error) {
	var err error

	c.balancesMapMux.RLock()
	bal, exists := c.balances[a]
	c.balancesMapMux.RUnlock()
	if !exists {
		bal, err = c.Client.BalanceAt(ctx, a, c.AtBlockNumber)
		if err != nil {
			return nil, err
		}
		c.balancesMapMux.Lock()
		c.balances[a] = bal
		c.balancesMapMux.Unlock()
	}
	return bal, nil
}

// SubBalance subtracts the value `diff` from the balance of account `a`.
// If no balance is cached yet, SubBalance will conduct a call to the ethereum
// node to get the state of the balance before modifying it.
// The modified value is then persisted in the internal state cache.
func (c *Cached) SubBalance(ctx context.Context, a common.Address, diff *big.Int) error {
	c.balancesLocks.Lock(a)
	defer c.balancesLocks.Unlock(a)
	return c.subBalance(ctx, a, diff)
}

func (c *Cached) subBalance(ctx context.Context, a common.Address, diff *big.Int) error {
	old, err := c.getBalance(ctx, a)
	if err != nil {
		return err
	}
	newBal := new(big.Int).Sub(old, diff)
	if newBal.Sign() == -1 {
		return errors.New("subtracted balance would be negative")
	}
	c.balancesMapMux.Lock()
	c.balances[a] = newBal
	c.balancesMapMux.Unlock()
	return nil
}

// AddBalance adds the value `diff` to the balance of account `a`.
// If no balance is cached yet, AddBalance will conduct a call to the ethereum
// node to get the state of the balance before modifying it.
// The modified value is then persisted in the internal state cache.
func (c *Cached) AddBalance(ctx context.Context, a common.Address, diff *big.Int) error {
	c.balancesLocks.Lock(a)
	defer c.balancesLocks.Unlock(a)
	return c.addBalance(ctx, a, diff)
}

func (c *Cached) addBalance(ctx context.Context, a common.Address, diff *big.Int) error {
	old, err := c.getBalance(ctx, a)
	if err != nil {
		return err
	}
	c.balancesMapMux.Lock()
	c.balances[a] = new(big.Int).Add(old, diff)
	c.balancesMapMux.Unlock()
	return nil
}

// GetNonce polls and caches the state of account `a` balance at the
// block number ChainBatchCache.AtBlockNumber.
func (c *Cached) GetNonce(ctx context.Context, a common.Address) (uint64, error) {
	c.noncesLocks.Lock(a)
	defer c.noncesLocks.Unlock(a)
	var (
		err   error
		nonce uint64
	)
	c.noncesMapMux.RLock()
	nonce, exists := c.nonces[a]
	c.noncesMapMux.RUnlock()
	if !exists {
		nonce, err = c.Client.NonceAt(ctx, a, c.AtBlockNumber)
		if err != nil {
			return nonce, err
		}
		c.noncesMapMux.Lock()
		c.nonces[a] = nonce
		c.noncesMapMux.Unlock()
	}
	return nonce, nil
}

// SetNonce sets the value in the nonce cache of account `a` to value `nonce`.
// Once this is set, a call to GetNonce() will not poll the node but simply
// return the set value.
func (c *Cached) SetNonce(a common.Address, nonce uint64) {
	c.noncesLocks.Lock(a)
	defer c.noncesLocks.Unlock(a)
	c.noncesMapMux.Lock()
	c.nonces[a] = nonce
	c.noncesMapMux.Unlock()
}

func (c *Cached) Purge(a common.Address) {
	c.noncesLocks.Lock(a)
	defer c.noncesLocks.Unlock(a)
	c.noncesMapMux.Lock()
	defer c.noncesMapMux.Unlock()

	c.balancesLocks.Lock(a)
	defer c.balancesLocks.Unlock(a)
	c.balancesMapMux.Lock()
	defer c.balancesMapMux.Unlock()

	delete(c.nonces, a)
	delete(c.balances, a)
}

var _ State = (*Cached)(nil)
