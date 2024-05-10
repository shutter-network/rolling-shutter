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
	fetchActivesAtSyncStart     bool
	privKey                     *ecdsa.PrivateKey

	handlerShutterState event.ShutterStateHandler
	handlerKeyperSet    event.KeyperSetHandler
	handlerEonPublicKey event.EonPublicKeyHandler
	handlerBlock        event.BlockHandler
}

func (o *options) verify() error {
	if o.clientURL != "" && o.ethClient != nil {
		return errors.New("'WithClient' and 'WithClientURL' options are mutually exclusive")
	}
	if o.clientURL == "" && o.ethClient == nil {
		return errors.New("either 'WithClient' or 'WithClientURL' options are expected")
	}
	// TODO: check for the existence of the contract addresses depending on
	// what handlers are not nil
	return nil
}

func (o *options) applyHandler(c *Client) error {
	var err error
	syncedServices := []syncer.ManualFilterHandler{}

	c.KeyperSetManager, err = bindings.NewKeyperSetManager(*o.keyperSetManagerAddress, o.ethClient)
	if err != nil {
		return err
	}
	c.kssync = &syncer.KeyperSetSyncer{
		Client:   o.ethClient,
		Contract: c.KeyperSetManager,
		Log:      c.log,
		Handler:  o.handlerKeyperSet,
	}
	if o.handlerKeyperSet != nil {
		syncedServices = append(syncedServices, c.kssync)
	}

	c.KeyBroadcast, err = bindings.NewKeyBroadcastContract(*o.keyBroadcastContractAddress, o.ethClient)
	if err != nil {
		return err
	}
	c.epksync = &syncer.EonPubKeySyncer{
		Client:           o.ethClient,
		Log:              c.log,
		KeyBroadcast:     c.KeyBroadcast,
		KeyperSetManager: c.KeyperSetManager,
		Handler:          o.handlerEonPublicKey,
	}
	if o.handlerEonPublicKey != nil {
		syncedServices = append(syncedServices, c.epksync)
	}
	c.sssync = &syncer.ShutterStateSyncer{
		Client:   o.ethClient,
		Contract: c.KeyperSetManager,
		Log:      c.log,
		Handler:  o.handlerShutterState,
	}
	if o.handlerShutterState != nil {
		syncedServices = append(syncedServices, c.sssync)
	}

	if o.handlerBlock == nil {
		// Even if the user is not interested in handling new block events,
		// the streaming block handler must be running in order to
		// synchronize polling of new contract events.
		// Since the handler function is always called, we need to
		// inject a noop-handler
		o.handlerBlock = func(ctx context.Context, lb *event.LatestBlock) error {
			return nil
		}
	}

	c.uhsync = &syncer.UnsafeHeadSyncer{
		Client:             o.ethClient,
		Log:                c.log,
		Handler:            o.handlerBlock,
		SyncedHandler:      syncedServices,
		FetchActiveAtStart: o.fetchActivesAtSyncStart,
		SyncStartBlock:     o.syncStart,
	}
	if o.handlerBlock != nil {
		c.services = append(c.services, c.uhsync)
	}
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

	// the nil passthrough will use "latest" for each call,
	// but we want to harmonize and fix the sync start to a specific block.
	if o.syncStart.IsLatest() {
		latestBlock, err := c.EthereumClient.BlockNumber(ctx)
		if err != nil {
			return errors.Wrap(err, "polling latest block")
		}
		o.syncStart = number.NewBlockNumber(&latestBlock)
	}

	if o.logger != nil {
		c.log = o.logger
	}

	c.privKey = o.privKey
	return o.applyHandler(c)
}

func defaultOptions() *options {
	return &options{
		keyperSetManagerAddress:     &predeploy.KeyperSetManagerAddr,
		keyBroadcastContractAddress: &predeploy.KeyBroadcastContractAddr,
		clientURL:                   "",
		ethClient:                   nil,
		logger:                      noopLogger,
		runner:                      nil,
		fetchActivesAtSyncStart:     true,
		syncStart:                   number.NewBlockNumber(nil),
	}
}

func WithNoFetchActivesBeforeStart() Option {
	return func(o *options) error {
		o.fetchActivesAtSyncStart = false
		return nil
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
