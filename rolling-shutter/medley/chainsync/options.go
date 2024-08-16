package chainsync

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/logger"
)

var noopLogger = &logger.NoopLogger{}

type Option func(*options) error

type options struct {
	clientURL      string
	ethClient      client.Sync
	logger         log.Logger
	syncStart      *number.BlockNumber
	blockCacheSize uint64
	eventHandler   []syncer.ContractEventHandler
	chainHandler   []syncer.ChainUpdateHandler
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

	//TODO: db chaincache when option supplied
	cache := syncer.NewMemoryChainCache(int(o.blockCacheSize), nil)
	f := syncer.NewFetcher(o.ethClient, cache, o.logger)

	return f, nil
}

func defaultOptions() *options {
	return &options{
		logger:         noopLogger,
		syncStart:      number.NewBlockNumber(nil),
		blockCacheSize: 50,
		eventHandler:   []syncer.ContractEventHandler{},
		chainHandler:   []syncer.ChainUpdateHandler{},
	}
}

func WithSyncStartBlock(
	blockNumber *number.BlockNumber,
) Option {
	if blockNumber == nil {
		blockNumber = number.NewBlockNumber(nil)
	}
	return func(o *options) error {
		o.syncStart = blockNumber
		return nil
	}
}

func WithClientURL(url string) Option {
	return func(o *options) error {
		o.clientURL = url
		return nil
	}
}

func WithBlockCacheSize(s uint64) Option {
	return func(o *options) error {
		o.blockCacheSize = s
		return nil
	}
}

func WithLogger(l log.Logger) Option {
	return func(o *options) error {
		o.logger = l
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
