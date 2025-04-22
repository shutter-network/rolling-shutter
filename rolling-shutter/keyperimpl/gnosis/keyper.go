package gnosis

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	gethLog "github.com/ethereum/go-ethereum/log"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	validatorRegistryBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/validatorregistry"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/eonkeypublisher"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/syncmonitor"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/beaconapiclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/slotticker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

var ErrParseKeyperSet = errors.New("cannot parse KeyperSet")

// The relative proposal timeout specifies for how long we wait for a block proposal to appear in
// a block. If we don't receive one in this time, we assume the slot is empty. The timeout is
// given as a fraction of the slot duration.
const (
	relativeProposalTimeoutNumerator   = 1
	relativeProposalTimeoutDenominator = 3
)

type Keyper struct {
	core            *keyper.KeyperCore
	config          *Config
	dbpool          *pgxpool.Pool
	beaconAPIClient *beaconapiclient.Client

	chainSyncClient     *chainsync.Client
	sequencerSyncer     *SequencerSyncer
	validatorSyncer     *ValidatorSyncer
	eonKeyPublisher     *eonkeypublisher.EonKeyPublisher
	latestTriggeredSlot *uint64
	syncMonitor         *syncmonitor.SyncMonitor

	// input events
	newBlocks        chan *syncevent.LatestBlock
	newKeyperSets    chan *syncevent.KeyperSet
	newEonPublicKeys chan keyper.EonPublicKey
	slotTicker       *slotticker.SlotTicker

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

	kpr.latestTriggeredSlot = nil

	offset := -(time.Duration(kpr.config.Gnosis.SecondsPerSlot) * time.Second) *
		(relativeProposalTimeoutDenominator - relativeProposalTimeoutNumerator) /
		relativeProposalTimeoutDenominator
	kpr.slotTicker = slotticker.NewSlotTicker(
		time.Duration(kpr.config.Gnosis.SecondsPerSlot*uint64(time.Second)),
		time.Unix(int64(kpr.config.Gnosis.GenesisSlotTimestamp), 0),
		offset,
	)

	kpr.dbpool, err = db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	kpr.beaconAPIClient, err = beaconapiclient.New(kpr.config.BeaconAPIURL)
	if err != nil {
		return errors.Wrap(err, "failed to initialize beacon API client")
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
		chainsync.WithClientURL(kpr.config.Gnosis.Node.EthereumURL),
		chainsync.WithKeyperSetManager(kpr.config.Gnosis.Contracts.KeyperSetManager),
		chainsync.WithKeyBroadcastContract(kpr.config.Gnosis.Contracts.KeyBroadcastContract),
		chainsync.WithSyncNewBlock(kpr.channelNewBlock),
		chainsync.WithSyncNewKeyperSet(kpr.channelNewKeyperSet),
		chainsync.WithPrivateKey(kpr.config.Gnosis.Node.PrivateKey.Key),
		chainsync.WithLogger(gethLog.NewLogger(slog.Default().Handler())),
	)
	if err != nil {
		return err
	}

	eonKeyPublisherClient, err := ethclient.DialContext(ctx, kpr.config.Gnosis.Node.EthereumURL)
	if err != nil {
		return errors.Wrapf(err, "failed to dial ethereum node at %s", kpr.config.Gnosis.Node.EthereumURL)
	}
	kpr.eonKeyPublisher, err = eonkeypublisher.NewEonKeyPublisher(
		kpr.dbpool,
		eonKeyPublisherClient,
		kpr.config.Gnosis.Contracts.KeyperSetManager,
		kpr.config.Gnosis.Node.PrivateKey.Key,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize eon key publisher")
	}

	err = kpr.initSequencerSyncer(ctx)
	if err != nil {
		return err
	}
	err = kpr.initValidatorSyncer(ctx)
	if err != nil {
		return err
	}

	// Set all transaction pointer ages to infinity. They will be reset to zero when the next
	// decryption keys arrive, telling us the agreed upon pointer value. Pointer values that are
	// not in the db yet are not affected. They will be initialized to zero when we first access
	// them. This is most importantly the case for the pointer value of not yet started eons.
	gnosisDB := database.New(kpr.dbpool)
	err = gnosisDB.ResetAllTxPointerAges(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to reset transaction pointer age")
	}

	kpr.syncMonitor = &syncmonitor.SyncMonitor{
		CheckInterval: time.Duration(kpr.config.Gnosis.SyncMonitorCheckInterval) * time.Second,
		SyncState: &GnosisSyncState{
			kpr.dbpool,
		},
	}

	runner.Go(func() error { return kpr.processInputs(ctx) })
	return runner.StartService(kpr.core, kpr.chainSyncClient, kpr.slotTicker, kpr.eonKeyPublisher)
}

func NewKeyper(kpr *Keyper, messagingMiddleware *MessagingMiddleware) (*keyper.KeyperCore, error) {
	core, err := keyper.New(
		&kprconfig.Config{
			InstanceID:           kpr.config.InstanceID,
			DatabaseURL:          kpr.config.DatabaseURL,
			HTTPEnabled:          kpr.config.HTTPEnabled,
			HTTPListenAddress:    kpr.config.HTTPListenAddress,
			P2P:                  kpr.config.P2P,
			Ethereum:             kpr.config.Gnosis.Node,
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
	return core, err
}

// initSequencerSycer initializes the sequencer syncer if the keyper is known to be a member of a
// keyper set. Otherwise, the syncer will only be initialized once such a keyper set is observed to
// be added, as only then we will know which eon(s) we are responsible for.
func (kpr *Keyper) initSequencerSyncer(ctx context.Context) error {
	client, err := ethclient.DialContext(ctx, kpr.config.Gnosis.Node.EthereumURL)
	if err != nil {
		return errors.Wrap(err, "failed to dial Ethereum execution node")
	}

	log.Info().
		Str("contract-address", kpr.config.Gnosis.Contracts.KeyperSetManager.Hex()).
		Msg("initializing sequencer syncer")
	contract, err := sequencerBindings.NewSequencer(kpr.config.Gnosis.Contracts.Sequencer, client)
	if err != nil {
		return err
	}
	kpr.sequencerSyncer = &SequencerSyncer{
		Contract:             contract,
		DBPool:               kpr.dbpool,
		ExecutionClient:      client,
		GenesisSlotTimestamp: kpr.config.Gnosis.GenesisSlotTimestamp,
		SecondsPerSlot:       kpr.config.Gnosis.SecondsPerSlot,
		SyncStartBlockNumber: kpr.config.Gnosis.SyncStartBlockNumber,
	}

	// Perform an initial sync now because it might take some time and doing so during regular
	// slot processing might hold up things
	latestHeader, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get latest block header")
	}
	err = kpr.sequencerSyncer.Sync(ctx, latestHeader)
	if err != nil {
		return err
	}

	return nil
}

func (kpr *Keyper) initValidatorSyncer(ctx context.Context) error {
	validatorSyncerClient, err := ethclient.DialContext(ctx, kpr.config.Gnosis.Node.EthereumURL)
	if err != nil {
		return errors.Wrap(err, "failed to dial ethereum node")
	}
	chainID, err := validatorSyncerClient.ChainID(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get chain ID")
	}
	validatorRegistryContract, err := validatorRegistryBindings.NewValidatorregistry(
		kpr.config.Gnosis.Contracts.ValidatorRegistry,
		validatorSyncerClient,
	)
	if err != nil {
		return errors.Wrap(err, "failed to instantiate validator registry contract")
	}
	kpr.validatorSyncer = &ValidatorSyncer{
		Contract:             validatorRegistryContract,
		DBPool:               kpr.dbpool,
		BeaconAPIClient:      kpr.beaconAPIClient,
		ExecutionClient:      validatorSyncerClient,
		ChainID:              chainID.Uint64(),
		SyncStartBlockNumber: kpr.config.Gnosis.SyncStartBlockNumber,
	}

	// Perform an initial sync now because it might take some time and doing so during regular
	// slot processing might hold up things
	latestHeader, err := validatorSyncerClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get latest block header")
	}
	err = kpr.validatorSyncer.Sync(ctx, latestHeader)
	if err != nil {
		return err
	}
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
		case slot := <-kpr.slotTicker.C:
			err = kpr.processNewSlot(ctx, slot)
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

func (kpr *Keyper) channelNewBlock(_ context.Context, ev *syncevent.LatestBlock) error {
	kpr.newBlocks <- ev
	return nil
}

func (kpr *Keyper) channelNewKeyperSet(_ context.Context, ev *syncevent.KeyperSet) error {
	kpr.newKeyperSets <- ev
	return nil
}

func (kpr *Keyper) channelNewEonPublicKey(_ context.Context, key keyper.EonPublicKey) error {
	kpr.newEonPublicKeys <- key
	return nil
}
