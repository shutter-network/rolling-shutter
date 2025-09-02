package primev

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/ethclient"
	gethLog "github.com/ethereum/go-ethereum/log"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/eonkeypublisher"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	providerregistry "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/primev/abi"
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

	chainSyncClient        *chainsync.Client
	providerRegistrySyncer *ProviderRegistrySyncer
	eonKeyPublisher        *eonkeypublisher.EonKeyPublisher
	newKeyperSets          chan *syncevent.KeyperSet
	newEonPublicKeys       chan keyper.EonPublicKey
	newBlocks              chan *syncevent.LatestBlock

	// outputs
	decryptionTriggerChannel chan *broker.Event[*epochkghandler.DecryptionTrigger]
}

func New(c *Config) *Keyper {
	return &Keyper{
		config: c,
	}
}

func (k *Keyper) Start(ctx context.Context, runner service.Runner) error {
	var err error

	k.newKeyperSets = make(chan *syncevent.KeyperSet)
	k.newEonPublicKeys = make(chan keyper.EonPublicKey)
	k.newBlocks = make(chan *syncevent.LatestBlock)
	k.decryptionTriggerChannel = make(chan *broker.Event[*epochkghandler.DecryptionTrigger])

	k.dbpool, err = db.Connect(ctx, runner, k.config.DatabaseURL, database.Definition.Name())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	messageSender, err := p2p.New(k.config.P2P)
	if err != nil {
		return errors.Wrap(err, "failed to initialize p2p messaging")
	}

	// TODO: do we need a middleware also here?
	messageSender.AddMessageHandler(&PrimevCommitmentHandler{
		config:                   k.config,
		decryptionTriggerChannel: k.decryptionTriggerChannel,
	})

	k.core, err = NewKeyper(k)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
	}

	k.chainSyncClient, err = chainsync.NewClient(
		ctx,
		chainsync.WithClientURL(k.config.Chain.Node.EthereumURL),
		chainsync.WithKeyperSetManager(k.config.Chain.Contracts.KeyperSetManager),
		chainsync.WithKeyBroadcastContract(k.config.Chain.Contracts.KeyBroadcastContract),
		chainsync.WithSyncNewKeyperSet(k.channelNewKeyperSet),
		chainsync.WithSyncNewBlock(k.channelNewBlock),
		chainsync.WithPrivateKey(k.config.Chain.Node.PrivateKey.Key),
		chainsync.WithLogger(gethLog.NewLogger(slog.Default().Handler())),
	)
	if err != nil {
		return err
	}

	eonKeyPublisherClient, err := ethclient.DialContext(ctx, k.config.Chain.Node.EthereumURL)
	if err != nil {
		return errors.Wrapf(err, "failed to dial ethereum node at %s", k.config.Chain.Node.EthereumURL)
	}
	k.eonKeyPublisher, err = eonkeypublisher.NewEonKeyPublisher(
		k.dbpool,
		eonKeyPublisherClient,
		k.config.Chain.Contracts.KeyperSetManager,
		k.config.Chain.Node.PrivateKey.Key,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize eon key publisher")
	}

	err = k.initRegistrySyncer(ctx)
	if err != nil {
		return err
	}

	runner.Go(func() error { return k.processInputs(ctx) })
	return runner.StartService(k.core, k.chainSyncClient, k.eonKeyPublisher)
}

func NewKeyper(kpr *Keyper) (*keyper.KeyperCore, error) {
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
	)
}

// initRegistrySyncer initializes the registry syncer.
func (k *Keyper) initRegistrySyncer(ctx context.Context) error {
	client, err := ethclient.DialContext(ctx, k.config.Primev.PrimevRPC)
	if err != nil {
		return errors.Wrap(err, "failed to dial Ethereum execution node")
	}

	log.Info().
		Str("contract-address", k.config.Chain.Contracts.KeyperSetManager.Hex()).
		Msg("initializing registry syncer")

	contract, err := providerregistry.NewProviderregistry(k.config.Primev.ProviderRegistryContract, client)
	if err != nil {
		return err
	}

	k.providerRegistrySyncer = &ProviderRegistrySyncer{
		Contract:             contract,
		DBPool:               k.dbpool,
		ExecutionClient:      client,
		SyncStartBlockNumber: k.config.Primev.SyncStartBlockNumber,
	}

	// Perform an initial sync now because it might take some time and doing so during regular
	// slot processing might hold up things
	latestHeader, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get latest block header")
	}
	err = k.providerRegistrySyncer.Sync(ctx, latestHeader)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keyper) processInputs(ctx context.Context) error {
	var err error
	for {
		select {
		case ev := <-k.newBlocks:
			err = k.processNewBlock(ctx, ev)
		case ev := <-k.newKeyperSets:
			err = k.processNewKeyperSet(ctx, ev)
		case ev := <-k.newEonPublicKeys:
			err = k.processNewEonPublicKey(ctx, ev)
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

func (k *Keyper) channelNewEonPublicKey(_ context.Context, key keyper.EonPublicKey) error {
	k.newEonPublicKeys <- key
	return nil
}

func (k *Keyper) channelNewKeyperSet(_ context.Context, ev *syncevent.KeyperSet) error {
	k.newKeyperSets <- ev
	return nil
}

func (k *Keyper) channelNewBlock(_ context.Context, ev *syncevent.LatestBlock) error {
	k.newBlocks <- ev
	return nil
}
