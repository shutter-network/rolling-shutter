package client

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	gethLog "github.com/ethereum/go-ethereum/log"
)

var ErrNotImplemented = errors.New("not implemented")

var _ Sync = &TestClient{}

type TestClient struct {
	log                    gethLog.Logger
	mux                    *sync.RWMutex
	subsMux                *sync.RWMutex
	headerChain            []*types.Header
	logs                   map[common.Hash][]types.Log
	latestHeadIndex        int
	initialProgress        bool
	latestHeadEmitter      []chan<- *types.Header
	latestHeadSubscription []*Subscription
}

func NewSubscription(idx int) *Subscription {
	return &Subscription{
		idx: idx,
		err: make(chan error, 1),
	}
}

type Subscription struct {
	idx int
	err chan error
}

func (su *Subscription) Unsubscribe() {
	// TODO: not implemented yet, but we don't want to panic
}

func (su *Subscription) Err() <-chan error {
	return su.err
}

type TestClientController struct {
	c *TestClient
}

func NewTestClient(logger gethLog.Logger) (*TestClient, *TestClientController) {
	c := &TestClient{
		log:                    log,
		mux:                    &sync.RWMutex{},
		subsMux:                &sync.RWMutex{},
		headerChain:            []*types.Header{},
		logs:                   map[common.Hash][]types.Log{},
		latestHeadIndex:        0,
		initialProgress:        false,
		latestHeadEmitter:      []chan<- *types.Header{},
		latestHeadSubscription: []*Subscription{},
	}
	ctrl := &TestClientController{c}
	return c, ctrl
}

// progresses the internal state of the latest head
// until no more information is available in the
// internal header chain.
func (c *TestClientController) ProgressAllHeads() {
	for c.ProgressHead() {
	}
}

// updates the internal state of the latest
// head one block. This will iterate over the
// internal headerChain and thus also includes reorging
// and decreasing the latest-head number.
func (c *TestClientController) ProgressHead() bool {
	c.c.mux.Lock()
	defer c.c.mux.Unlock()

	if c.c.latestHeadIndex >= len(c.c.headerChain)-1 {
		return false
	}
	c.c.latestHeadIndex++
	return true
}

func (c *TestClientController) WaitSubscribed(ctx context.Context) {
	for {
		c.c.subsMux.RLock()
		if len(c.c.latestHeadEmitter) > 0 {
			c.c.subsMux.RUnlock()
			break
		}
		c.c.subsMux.RUnlock()
		if ctx.Err() != nil {
			return
		}
		time.After(50 * time.Millisecond)
	}
}
func (c *TestClientController) EmitLatestHead(ctx context.Context) error {
	c.c.subsMux.RLock()
	defer c.c.subsMux.RUnlock()

	c.c.mux.RLock()
	if len(c.c.latestHeadEmitter) == 0 {
		c.c.mux.RUnlock()
		return nil
	}
	h := c.c.getLatestHeader()
	c.c.mux.RUnlock()
	for _, em := range c.c.latestHeadEmitter {
		select {
		case em <- h:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (c *TestClientController) AppendNextHeader(h *types.Header, events ...types.Log) {
	c.c.mux.Lock()
	defer c.c.mux.Unlock()

	c.c.headerChain = append(c.c.headerChain, h)
	_, ok := c.c.logs[h.Hash()]
	if ok {
		return
	}
	c.c.logs[h.Hash()] = events
}

func (t *TestClient) ChainID(_ context.Context) (*big.Int, error) { //nolint: unparam
	return big.NewInt(42), nil
}

func (t *TestClient) Close() {
	// TODO: cleanup
}

func (t *TestClient) getLatestHeader() *types.Header {
	if len(t.headerChain) == 0 {
		return nil
	}
	return t.headerChain[t.latestHeadIndex]
}

func (t *TestClient) searchBlock(f func(*types.Header) bool) *types.Header {
	for i := t.latestHeadIndex; i >= 0; i-- {
		h := t.headerChain[i]
		if f(h) {
			return h
		}
	}
	return nil
}

func (t *TestClient) searchBlockByNumber(number *big.Int) *types.Header {
	return t.searchBlock(
		func(h *types.Header) bool {
			return h.Number.Cmp(number) == 0
		})
}

func (t *TestClient) searchBlockByHash(hash common.Hash) *types.Header {
	return t.searchBlock(
		func(h *types.Header) bool {
			return hash.Cmp(h.Hash()) == 0
		})
}

func (t *TestClient) BlockNumber(_ context.Context) (uint64, error) { //nolint: unparam
	t.mux.RLock()
	defer t.mux.RUnlock()

	return t.getLatestHeader().Nonce.Uint64(), nil
}

func (t *TestClient) HeaderByHash(_ context.Context, hash common.Hash) (*types.Header, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()

	h := t.searchBlockByHash(hash)
	if h == nil {
		return nil, errors.New("header not found")
	}
	return h, nil
}

func (t *TestClient) HeaderByNumber(_ context.Context, number *big.Int) (*types.Header, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()

	if number == nil {
		return t.getLatestHeader(), nil
	}
	if number.Cmp(big.NewInt(-2)) == 0 {
		return t.getLatestHeader(), nil
	}
	h := t.searchBlockByNumber(number)
	if h == nil {
		return nil, errors.New("not found")
	}
	return h, nil
}

func (t *TestClient) SubscribeNewHead(_ context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	t.subsMux.Lock()
	defer t.subsMux.Unlock()

	t.latestHeadEmitter = append(t.latestHeadEmitter, ch)
	su := NewSubscription(len(t.latestHeadSubscription) - 1)
	t.latestHeadSubscription = append(t.latestHeadSubscription, su)
	// TODO: unsubscribe and deleting from the array
	// TODO: filling error promise in the subscription
	return su, nil
}

func (t *TestClient) getLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	logs := []types.Log{}
	if query.BlockHash != nil {
		log, ok := t.logs[*query.BlockHash]
		if !ok {
			// TODO: if possible return the same error as the client
			return logs, fmt.Errorf("no logs found")
		}
		return log, nil
	}
	if query.FromBlock != nil {
		current := query.FromBlock
		toBlock := query.ToBlock
		if toBlock == nil {
			latest := t.getLatestHeader()
			toBlock = latest.Number
		}
		for current.Cmp(toBlock) != +1 {
			h := t.searchBlockByNumber(current)

			current = new(big.Int).Add(current, big.NewInt(1))
			log, ok := t.logs[h.Hash()]
			if !ok {
				continue
			}
			logs = append(logs, log...)
		}
	}
	//FIXME: also return no logs found if empty?
	return logs, nil
}

func (t *TestClient) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()

	logs, err := t.getLogs(ctx, query)
	if len(logs) > 0 {
		t.log.Info("logs found in FilterLogs", "logs", logs)
	}
	if err != nil {
		return logs, err
	}
	filtered := []types.Log{}

	addrs := map[common.Address]struct{}{}
	for _, a := range query.Addresses {
		addrs[a] = struct{}{}
	}
	t.log.Info("query Addresses FilterLogs", "addresses", query.Addresses)

	for _, log := range logs {
		if _, ok := addrs[log.Address]; !ok {
			continue
		}
		filtered = append(filtered, log)
	}
	// OPTIM: filter by the topics, but this gets complex
	// since it's position based as well.
	// It's not strictly needed for the tests, since the downstream
	// caller should also ignore wrong log types upon parsing.
	return filtered, nil
}

func (t *TestClient) SubscribeFilterLogs(_ context.Context, _ ethereum.FilterQuery, _ chan<- types.Log) (ethereum.Subscription, error) {
	panic(ErrNotImplemented)
}

func (t *TestClient) CodeAt(_ context.Context, _ common.Address, _ *big.Int) ([]byte, error) {
	panic(ErrNotImplemented)
}

func (t *TestClient) TransactionReceipt(_ context.Context, _ common.Hash) (*types.Receipt, error) {
	panic(ErrNotImplemented)
}

func (t *TestClient) BlockByHash(_ context.Context, _ common.Hash) (*types.Block, error) {
	panic(ErrNotImplemented)
}

func (t *TestClient) TransactionCount(_ context.Context, _ common.Hash) (uint, error) {
	panic(ErrNotImplemented)

}

func (t *TestClient) TransactionInBlock(_ context.Context, _ common.Hash, _ uint) (*types.Transaction, error) {
	panic(ErrNotImplemented)

}
