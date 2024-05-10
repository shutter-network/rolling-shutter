package client

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var ErrNotImplemented = errors.New("not implemented")

var _ SyncEthereumClient = &TestClient{}

type TestClient struct {
	headerChain            []*types.Header
	latestHeadIndex        int
	intialProgress         bool
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

func NewTestClient() (*TestClient, *TestClientController) {
	c := &TestClient{
		headerChain:     []*types.Header{},
		latestHeadIndex: 0,
	}
	ctrl := &TestClientController{c}
	return c, ctrl
}

func (c *TestClientController) ProgressHead() bool {
	if c.c.latestHeadIndex >= len(c.c.headerChain)-1 {
		return false
	}
	c.c.latestHeadIndex++
	return true
}

func (c *TestClientController) EmitEvents(ctx context.Context) error {
	if len(c.c.latestHeadEmitter) == 0 {
		return nil
	}
	h := c.c.getLatestHeader()
	for _, em := range c.c.latestHeadEmitter {
		select {
		case em <- h:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (c *TestClientController) AppendNextHeaders(h ...*types.Header) {
	c.c.headerChain = append(c.c.headerChain, h...)
}

func (t *TestClient) ChainID(_ context.Context) (*big.Int, error) {
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

func (t *TestClient) BlockNumber(_ context.Context) (uint64, error) {
	return t.getLatestHeader().Nonce.Uint64(), nil
}

func (t *TestClient) HeaderByHash(_ context.Context, hash common.Hash) (*types.Header, error) {
	h := t.searchBlockByHash(hash)
	if h == nil {
		return nil, errors.New("header not found")
	}
	return h, nil
}

func (t *TestClient) HeaderByNumber(_ context.Context, number *big.Int) (*types.Header, error) {
	if number == nil {
		return t.getLatestHeader(), nil
	}
	h := t.searchBlockByNumber(number)
	if h == nil {
		return nil, errors.New("header not found")
	}
	return h, nil
}

func (t *TestClient) SubscribeNewHead(_ context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	t.latestHeadEmitter = append(t.latestHeadEmitter, ch)
	su := NewSubscription(len(t.latestHeadSubscription) - 1)
	t.latestHeadSubscription = append(t.latestHeadSubscription, su)
	// TODO: unsubscribe and deleting from the array
	// TODO: filling error promise in the subscription
	return su, nil
}

func (t *TestClient) FilterLogs(_ context.Context, _ ethereum.FilterQuery) ([]types.Log, error) {
	panic(ErrNotImplemented)
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
