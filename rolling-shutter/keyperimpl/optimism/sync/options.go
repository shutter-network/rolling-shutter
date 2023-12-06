package sync

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shutter-network/shop-contracts/bindings"
	"github.com/shutter-network/shop-contracts/predeploy"

	syncclient "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type Option func(*options) error

type options struct {
	keyperSetManagerAddress     *common.Address
	keyBroadcastContractAddress *common.Address
	clientURL                   string
	client                      syncclient.Client
	logger                      log.Logger
	runner                      service.Runner
	syncStart                   *uint64

	handlerShutterState event.ShutterStateHandler
	handlerKeyperSet    event.KeyperSetHandler
	handlerEonPublicKey event.EonPublicKeyHandler
	handlerBlock        event.BlockHandler
}

func (o *options) verify() error {
	if o.clientURL != "" && o.client != nil {
		// TODO: error message
		return errors.New("can't use client and client url")
	}
	if o.clientURL == "" && o.client == nil {
		// TODO: error message
		return errors.New("have to provide either url or client")
	}
	// TODO: check for the existence of the contract addresses depending on
	// what handlers are not nil
	return nil
}

// initialize the shutter client and apply the options.
// the context is only the initialisation context,
// and should not be considered to handle the lifecycle
// of shutter clients background workers.
func (o *options) apply(ctx context.Context, c *ShutterL2Client) error {
	var (
		client syncclient.Client
		err    error
	)
	if o.clientURL != "" {
		o.client, err = ethclient.DialContext(ctx, o.clientURL)
		if err != nil {
			return err
		}
	}
	client = o.client
	c.log = o.logger

	c.Client = client

	c.KeyperSetManager, err = bindings.NewKeyperSetManager(*o.keyperSetManagerAddress, client)
	if err != nil {
		return err
	}
	c.kssync = &syncer.KeyperSetSyncer{
		Client:     client,
		Contract:   c.KeyperSetManager,
		Log:        c.log,
		StartBlock: o.syncStart,
		Handler:    o.handlerKeyperSet,
	}
	if o.handlerKeyperSet != nil {
		c.services = append(c.services, c.kssync)
	}

	c.KeyBroadcast, err = bindings.NewKeyBroadcastContract(*o.keyBroadcastContractAddress, client)
	if err != nil {
		return err
	}
	c.epksync = &syncer.EonPubKeySyncer{
		Client:     client,
		Log:        c.log,
		Contract:   c.KeyBroadcast,
		Handler:    o.handlerEonPublicKey,
		StartBlock: o.syncStart,
	}
	if o.handlerEonPublicKey != nil {
		c.services = append(c.services, c.epksync)
	}

	c.sssync = &syncer.ShutterStateSyncer{
		Client:     client,
		Contract:   c.KeyperSetManager,
		Log:        c.log,
		Handler:    o.handlerShutterState,
		StartBlock: o.syncStart,
	}
	if o.handlerShutterState != nil {
		c.services = append(c.services, c.sssync)
	}

	if o.handlerBlock != nil {
		c.uhsync = &syncer.UnsafeHeadSyncer{
			Client:  client,
			Log:     c.log,
			Handler: o.handlerBlock,
		}
	}
	return nil
}

func defaultOptions() *options {
	return &options{
		keyperSetManagerAddress:     &predeploy.KeyperSetManagerAddr,
		keyBroadcastContractAddress: &predeploy.KeyBroadcastContractAddr,
		clientURL:                   "",
		client:                      nil,
		logger:                      noopLogger,
		runner:                      nil,
		syncStart:                   nil,
	}
}

func WithSyncStartBlock(blockNumber uint64) Option {
	return func(o *options) error {
		bn := blockNumber
		o.syncStart = &bn
		return nil
	}
}

func WithRunner(runner service.Runner) Option {
	return func(o *options) error {
		o.runner = runner
		return nil
	}
}

func WithKeyBroadcastContract(address common.Address) Option {
	return func(o *options) error {
		o.keyBroadcastContractAddress = &address
		return nil
	}
}

func WithKeyperSetManager(address common.Address) Option {
	return func(o *options) error {
		o.keyperSetManagerAddress = &address
		return nil
	}
}

func WithClientURL(url string) Option {
	return func(o *options) error {
		o.clientURL = url
		return nil
	}
}

func WithLogger(l log.Logger) Option {
	return func(o *options) error {
		o.logger = l
		return nil
	}
}

func WithClient(client syncclient.Client) Option {
	return func(o *options) error {
		o.client = client
		return nil
	}
}

func WithSyncNewKeyperSet(handler event.KeyperSetHandler) Option {
	return func(o *options) error {
		o.handlerKeyperSet = handler
		return nil
	}
}

func WithSyncNewBlock(handler event.BlockHandler) Option {
	return func(o *options) error {
		o.handlerBlock = handler
		return nil
	}
}

func WithSyncNewEonKey(handler event.EonPublicKeyHandler) Option {
	return func(o *options) error {
		o.handlerEonPublicKey = handler
		return nil
	}
}

func WithSyncNewShutterState(handler event.ShutterStateHandler) Option {
	return func(o *options) error {
		o.handlerShutterState = handler
		return nil
	}
}