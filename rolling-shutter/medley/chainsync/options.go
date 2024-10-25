package chainsync

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

const defaultMemoryBlockCacheSize = 50

type Option func(*options) error

type options struct {
	clientURL    string
	ethClient    client.Sync
	syncStart    *number.BlockNumber
	chainCache   syncer.ChainCache
	eventHandler []syncer.ContractEventHandler
	chainHandler []syncer.ChainUpdateHandler
}

func (o *options) verify() error {
	if o.clientURL != "" && o.ethClient != nil {
		return errors.New("'WithClient' and 'WithClientURL' options are mutually exclusive")
	}
	if o.clientURL == "" && o.ethClient == nil {
		return errors.New("either 'WithClient' or 'WithClientURL' options are expected")
	}
	return nil
}

// initFetcher applies the options and initializes the fetcher.
// The context is only the initialisation context,
// and should not be considered to handle the lifecycle
// of shutter clients background workers.
func (o *options) initFetcher(ctx context.Context) (*syncer.Fetcher, error) {
	var err error
	if o.clientURL != "" {
		o.ethClient, err = ethclient.DialContext(ctx, o.clientURL)
		if err != nil {
			return nil, err
		}
	}

	if o.chainCache == nil {
		o.chainCache = syncer.NewMemoryChainCache(int(defaultMemoryBlockCacheSize), nil)
	}
	f := syncer.NewFetcher(o.ethClient, o.chainCache)

	for _, h := range o.chainHandler {
		f.RegisterChainUpdateHandler(h)
	}
	for _, h := range o.eventHandler {
		f.RegisterContractEventHandler(h)
	}
	return f, nil
}

func defaultOptions() *options {
	return &options{
		syncStart:    number.NewBlockNumber(nil),
		eventHandler: []syncer.ContractEventHandler{},
		chainHandler: []syncer.ChainUpdateHandler{},
	}
}

func WithClientURL(url string) Option {
	return func(o *options) error {
		o.clientURL = url
		return nil
	}
}

// NOTE: The Latest() of the chaincache determines what is the starting
// point of the chainsync.
// In case of an empty chaincache, we will initialize the cache
// with the current latest block.
// If we have a very old (persistent) chaincache, we will sync EVERY block
// since the latest known block of the cache due to consistency considerations.
// If that is unfeasible, the cache has to be emptied beforehand and the
// gap in state-updates has to be dealt with or accepted.
// If NO chaincache is passed with this option, an empty in-memory
// chain-cache with a capped cachesize will be used.
func WithChainCache(c syncer.ChainCache) Option {
	return func(o *options) error {
		o.chainCache = c
		return nil
	}
}

func WithClient(c client.Sync) Option {
	return func(o *options) error {
		o.ethClient = c
		return nil
	}
}

func WithContractEventHandler(h syncer.ContractEventHandler) Option {
	return func(o *options) error {
		o.eventHandler = append(o.eventHandler, h)
		return nil
	}
}

func WithChainUpdateHandler(h syncer.ChainUpdateHandler) Option {
	return func(o *options) error {
		o.chainHandler = append(o.chainHandler, h)
		return nil
	}
}
