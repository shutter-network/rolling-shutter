package chainsync

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"
	"github.com/shutter-network/shop-contracts/predeploy"

	syncclient "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type Option func(*options) error

type options struct {
	keyperSetManagerAddress     *common.Address
	keyBroadcastContractAddress *common.Address
	clientURL                   string
	ethClient                   syncclient.EthereumClient
	logger                      log.Logger
	runner                      service.Runner
	syncStart                   *number.BlockNumber
	privKey                     *ecdsa.PrivateKey

	handlerShutterState event.ShutterStateHandler
	handlerKeyperSet    event.KeyperSetHandler
	handlerEonPublicKey event.EonPublicKeyHandler
	handlerBlock        event.BlockHandler
}

func (o *options) verify() error {
	if o.clientURL != "" && o.ethClient != nil {
		// TODO: error message
		return errors.New("can't use client and client url")
	}
	if o.clientURL == "" && o.ethClient == nil {
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
func (o *options) apply(ctx context.Context, c *Client) error {
	var (
		client syncclient.EthereumClient
		err    error
	)
	if o.clientURL != "" {
		o.ethClient, err = ethclient.DialContext(ctx, o.clientURL)
		if err != nil {
			return err
		}
	}
	client = o.ethClient

	c.EthereumClient = client

	if o.logger != nil {
		c.log = o.logger
		// NOCHECKIN:
		c.log.Info("got logger in options")
	}

	syncedServices := []syncer.ManualFilterHandler{}
	// the nil passthrough will use "latest" for each call,
	// but we want to harmonize and fix the sync start to a specific block.
	if o.syncStart.IsLatest() {
		latestBlock, err := c.EthereumClient.BlockNumber(ctx)
		if err != nil {
			return errors.Wrap(err, "polling latest block")
		}
		o.syncStart = number.NewBlockNumber(&latestBlock)
	}

	c.KeyperSetManager, err = bindings.NewKeyperSetManager(*o.keyperSetManagerAddress, client)
	if err != nil {
		return err
	}
	c.kssync = &syncer.KeyperSetSyncer{
		Client:              client,
		Contract:            c.KeyperSetManager,
		Log:                 c.log,
		StartBlock:          o.syncStart,
		Handler:             o.handlerKeyperSet,
		DisableEventWatcher: true,
	}
	if o.handlerKeyperSet != nil {
		c.services = append(c.services, c.kssync)
		syncedServices = append(syncedServices, c.kssync)
	}

	c.KeyBroadcast, err = bindings.NewKeyBroadcastContract(*o.keyBroadcastContractAddress, client)
	if err != nil {
		return err
	}
	c.epksync = &syncer.EonPubKeySyncer{
		Client:              client,
		Log:                 c.log,
		KeyBroadcast:        c.KeyBroadcast,
		KeyperSetManager:    c.KeyperSetManager,
		Handler:             o.handlerEonPublicKey,
		StartBlock:          o.syncStart,
		DisableEventWatcher: true,
	}
	if o.handlerEonPublicKey != nil {
		c.services = append(c.services, c.epksync)
		syncedServices = append(syncedServices, c.kssync)
	}

	c.sssync = &syncer.ShutterStateSyncer{
		Client:              client,
		Contract:            c.KeyperSetManager,
		Log:                 c.log,
		Handler:             o.handlerShutterState,
		StartBlock:          o.syncStart,
		DisableEventWatcher: true,
	}
	if o.handlerShutterState != nil {
		c.services = append(c.services, c.sssync)
		syncedServices = append(syncedServices, c.kssync)
	}

	if o.handlerBlock == nil {
		// NOOP - but we need to run the UnsafeHeadSyncer.
		// This is to keep the inner workings consisten,
		// we use the DisableEventWatcher mechanism in combination
		// with the UnsafeHeadSyncer instead of the streaming
		// Watch... subscription on events.
		// TODO: think about allowing the streaming events,
		// when guaranteed event order (event1,event2,new-block-event)
		// is not required
		o.handlerBlock = func(ctx context.Context, lb *event.LatestBlock) error {
			return nil
		}
	}

	c.uhsync = &syncer.UnsafeHeadSyncer{
		Client:        client,
		Log:           c.log,
		Handler:       o.handlerBlock,
		SyncedHandler: syncedServices,
	}
	if o.handlerBlock != nil {
		c.services = append(c.services, c.uhsync)
	}
	c.privKey = o.privKey
	return nil
}

func defaultOptions() *options {
	return &options{
		keyperSetManagerAddress:     &predeploy.KeyperSetManagerAddr,
		keyBroadcastContractAddress: &predeploy.KeyBroadcastContractAddr,
		clientURL:                   "",
		ethClient:                   nil,
		logger:                      noopLogger,
		runner:                      nil,
		syncStart:                   number.NewBlockNumber(nil),
	}
}

func WithSyncStartBlock(blockNumber *number.BlockNumber) Option {
	if blockNumber == nil {
		blockNumber = number.NewBlockNumber(nil)
	}
	return func(o *options) error {
		o.syncStart = blockNumber
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

func WithClient(client syncclient.EthereumClient) Option {
	return func(o *options) error {
		o.ethClient = client
		return nil
	}
}

func WithPrivateKey(key *ecdsa.PrivateKey) Option {
	return func(o *options) error {
		o.privKey = key
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
