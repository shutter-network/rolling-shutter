package shutterservice

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	gethLog "github.com/ethereum/go-ethereum/log"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	registryBindings "github.com/shutter-network/contracts/v2/bindings/shutterregistry"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/eonkeypublisher"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	eventTriggerRegistryBindings "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/help"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

var ErrParseKeyperSet = errors.New("cannot parse KeyperSet")

type Keyper struct {
	core   *keyper.KeyperCore
	config *Config
	dbpool *pgxpool.Pool

	chainSyncClient     *chainsync.Client
	registrySyncer      *RegistrySyncer
	eonKeyPublisher     *eonkeypublisher.EonKeyPublisher
	latestTriggeredTime *uint64
	syncMonitor         *SyncMonitor
	multiEventSyncer    *MultiEventSyncer

	// input events
	newBlocks        chan *syncevent.LatestBlock
	newKeyperSets    chan *syncevent.KeyperSet
	newEonPublicKeys chan keyper.EonPublicKey

	// outputs
	decryptionTriggerChannel chan *broker.Event[*epochkghandler.DecryptionTrigger]
}

func New(c *Config) *Keyper {
	return &Keyper{
		config: c,
	}
}

func (kpr *Keyper) Start(ctx context.Context, runner service.Runner) error {
	var err error

	kpr.newBlocks = make(chan *syncevent.LatestBlock)
	kpr.newKeyperSets = make(chan *syncevent.KeyperSet)
	kpr.newEonPublicKeys = make(chan keyper.EonPublicKey)
	kpr.decryptionTriggerChannel = make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	kpr.latestTriggeredTime = nil

	kpr.dbpool, err = db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	messageSender, err := p2p.New(kpr.config.P2P)
	if err != nil {
		return errors.Wrap(err, "failed to initialize p2p messaging")
	}

	messageSender.AddMessageHandler(&DecryptionKeySharesHandler{kpr.dbpool})
	messageSender.AddMessageHandler(&DecryptionKeysHandler{kpr.dbpool})
	messagingMiddleware := NewMessagingMiddleware(messageSender, kpr.dbpool, kpr.config)

	kpr.core, err = NewKeyper(kpr, messagingMiddleware)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
	}
	kpr.chainSyncClient, err = chainsync.NewClient(
		ctx,
		chainsync.WithClientURL(kpr.config.Chain.Node.EthereumURL),
		chainsync.WithKeyperSetManager(kpr.config.Chain.Contracts.KeyperSetManager),
		chainsync.WithKeyBroadcastContract(kpr.config.Chain.Contracts.KeyBroadcastContract),
		chainsync.WithSyncNewBlock(kpr.channelNewBlock),
		chainsync.WithSyncNewKeyperSet(kpr.channelNewKeyperSet),
		chainsync.WithPrivateKey(kpr.config.Chain.Node.PrivateKey.Key),
		chainsync.WithLogger(gethLog.NewLogger(slog.Default().Handler())),
	)
	if err != nil {
		return err
	}

	eonKeyPublisherClient, err := ethclient.DialContext(ctx, kpr.config.Chain.Node.EthereumURL)
	if err != nil {
		return errors.Wrapf(err, "failed to dial ethereum node at %s", kpr.config.Chain.Node.EthereumURL)
	}
	kpr.eonKeyPublisher, err = eonkeypublisher.NewEonKeyPublisher(
		kpr.dbpool,
		eonKeyPublisherClient,
		kpr.config.Chain.Contracts.KeyperSetManager,
		kpr.config.Chain.Node.PrivateKey.Key,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize eon key publisher")
	}

	err = kpr.initRegistrySyncer(ctx)
	if err != nil {
		return err
	}

	err = kpr.initMultiEventSyncer(ctx)
	if err != nil {
		return err
	}

	kpr.syncMonitor = &SyncMonitor{
		DBPool:        kpr.dbpool,
		CheckInterval: time.Duration(kpr.config.Chain.SyncMonitorCheckInterval) * time.Second,
	}
	runner.Go(func() error { return kpr.processInputs(ctx) })
	return runner.StartService(kpr.core, kpr.chainSyncClient, kpr.eonKeyPublisher, kpr.syncMonitor)
}

func NewKeyper(kpr *Keyper, messagingMiddleware *MessagingMiddleware) (*keyper.KeyperCore, error) {
	return keyper.New(
		&kprconfig.Config{
			InstanceID:           kpr.config.InstanceID,
			DatabaseURL:          kpr.config.DatabaseURL,
			HTTPEnabled:          kpr.config.HTTPEnabled,
			HTTPReadOnly:         kpr.config.HTTPReadOnly,
			HTTPListenAddress:    kpr.config.HTTPListenAddress,
			P2P:                  kpr.config.P2P,
			Ethereum:             kpr.config.Chain.Node,
			Shuttermint:          kpr.config.Shuttermint,
			Metrics:              kpr.config.Metrics,
			MaxNumKeysPerMessage: kpr.config.MaxNumKeysPerMessage,
		},
		kpr.decryptionTriggerChannel,
		keyper.WithDBPool(kpr.dbpool),
		keyper.NoBroadcastEonPublicKey(),
		keyper.WithEonPublicKeyHandler(kpr.channelNewEonPublicKey),
		keyper.WithMessaging(messagingMiddleware),
	)
}

// initRegistrySycer initializes the registry syncer if the keyper is known to be a member of a
// keyper set. Otherwise, the syncer will only be initialized once such a keyper set is observed to
// be added, as only then we will know which eon(s) we are responsible for.
func (kpr *Keyper) initRegistrySyncer(ctx context.Context) error {
	client, err := ethclient.DialContext(ctx, kpr.config.Chain.Node.EthereumURL)
	if err != nil {
		return errors.Wrap(err, "failed to dial Ethereum execution node")
	}

	log.Info().
		Str("contract-address", kpr.config.Chain.Contracts.KeyperSetManager.Hex()).
		Msg("initializing registry syncer")

	contract, err := registryBindings.NewShutterregistry(kpr.config.Chain.Contracts.ShutterRegistry, client)
	if err != nil {
		return err
	}

	// TODO: need to update go module after contract is finalized
	kpr.registrySyncer = &RegistrySyncer{
		Contract:             contract,
		DBPool:               kpr.dbpool,
		ExecutionClient:      client,
		SyncStartBlockNumber: kpr.config.Chain.SyncStartBlockNumber,
	}

	// Perform an initial sync now because it might take some time and doing so during regular
	// slot processing might hold up things
	latestHeader, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get latest block header")
	}
	err = kpr.registrySyncer.Sync(ctx, latestHeader)
	if err != nil {
		return err
	}

	return nil
}

// initMultiEventSyncer initializes the multi event syncer and all its event processors.
func (kpr *Keyper) initMultiEventSyncer(ctx context.Context) error {
	triggerRegistryClient, err := ethclient.DialContext(ctx, kpr.config.Chain.Node.EthereumURL)
	if err != nil {
		return fmt.Errorf("failed to dial Ethereum execution node: %w", err)
	}
	eventTriggerRegistryContract, err := eventTriggerRegistryBindings.NewShutterRegistry(kpr.config.Chain.Contracts.EventTriggerRegistry, triggerRegistryClient)
	if err != nil {
		return fmt.Errorf("failed to create ShutterRegistry contract instance: %w", err)
	}
	eventTriggerRegisteredProcessor := NewEventTriggerRegisteredEventProcessor(
		eventTriggerRegistryContract,
		kpr.dbpool,
	)

	triggerClient, err := ethclient.DialContext(ctx, kpr.config.Chain.Node.EthereumURL)
	if err != nil {
		return fmt.Errorf("failed to dial Ethereum execution node: %w", err)
	}
	triggerProcessor := NewTriggerProcessor(triggerClient, kpr.dbpool)

	processors := []EventProcessor{
		eventTriggerRegisteredProcessor,
		triggerProcessor,
	}

	multiEventSyncerClient, err := ethclient.DialContext(ctx, kpr.config.Chain.Node.EthereumURL)
	if err != nil {
		return fmt.Errorf("failed to dial Ethereum node at %s: %w", kpr.config.Chain.Node.EthereumURL, err)
	}
	kpr.multiEventSyncer, err = NewMultiEventSyncer(
		kpr.dbpool,
		multiEventSyncerClient,
		kpr.config.Chain.SyncStartBlockNumber,
		processors,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize multi event syncer: %w", err)
	}

	// Perform an initial sync now because it might take some time and doing so during regular
	// slot processing might hold up things
	log.Info().Msg("performing initial sync of multi event syncer")
	latestHeader, err := multiEventSyncerClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get latest block header: %w", err)
	}
	err = kpr.multiEventSyncer.Sync(ctx, latestHeader)
	if err != nil {
		return fmt.Errorf("failed to perform initial sync: %w", err)
	}
	log.Info().Msg("multi event syncer initialized")

	return nil
}

func (kpr *Keyper) processInputs(ctx context.Context) error {
	var err error
	for {
		select {
		case ev := <-kpr.newBlocks:
			err = kpr.processNewBlock(ctx, ev)
		case ev := <-kpr.newKeyperSets:
			err = kpr.processNewKeyperSet(ctx, ev)
		case ev := <-kpr.newEonPublicKeys:
			err = kpr.processNewEonPublicKey(ctx, ev)
		case <-ctx.Done():
			return ctx.Err()
		}
		if err != nil {
			// TODO: Check if it's safe to drop those events. If not, we should store the
			// ones that remain on the channel in the db and process them when we restart.
			// TODO: also, should we stop the keyper or just log the error and continue?
			// return err
			log.Error().Err(err).Msg("error processing event")
		}
	}
}

func (kpr *Keyper) channelNewEonPublicKey(_ context.Context, key keyper.EonPublicKey) error {
	kpr.newEonPublicKeys <- key
	return nil
}

func (kpr *Keyper) channelNewBlock(_ context.Context, ev *syncevent.LatestBlock) error {
	kpr.newBlocks <- ev
	return nil
}

func (kpr *Keyper) channelNewKeyperSet(_ context.Context, ev *syncevent.KeyperSet) error {
	kpr.newKeyperSets <- ev
	return nil
}
