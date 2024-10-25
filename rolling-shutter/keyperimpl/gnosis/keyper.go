package gnosis

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/eonkeypublisher"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/synchandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/beaconapiclient"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/errs"
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
	config          *config.Config
	dbpool          *pgxpool.Pool
	beaconAPIClient *beaconapiclient.Client
	gnosisEthClient *ethclient.Client

	eonKeyPublisher     *eonkeypublisher.EonKeyPublisher
	latestTriggeredSlot *uint64

	newEonPublicKeys chan keyper.EonPublicKey
	slotTicker       *slotticker.SlotTicker

	// outputs
	decryptionTriggerChannel chan *broker.Event[*epochkghandler.DecryptionTrigger]
}

func New(c *config.Config) *Keyper {
	return &Keyper{
		config: c,
	}
}

func (kpr *Keyper) Start(ctx context.Context, runner service.Runner) error {
	var err error

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

	messaging, err := p2p.New(kpr.config.P2P)
	if err != nil {
		return errors.Wrap(err, "failed to initialize p2p messaging")
	}
	messaging.AddMessageHandler(&DecryptionKeySharesHandler{kpr.dbpool})
	messaging.AddMessageHandler(&DecryptionKeysHandler{kpr.dbpool})

	chainID, err := kpr.gnosisEthClient.ChainID(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get chain ID")
	}

	kpr.gnosisEthClient, err = ethclient.DialContext(ctx, kpr.config.Gnosis.Node.EthereumURL)
	if err != nil {
		return errors.Wrapf(err, "failed to dial gnosis node at %s", kpr.config.Gnosis.Node.EthereumURL)
	}
	kpr.eonKeyPublisher, err = eonkeypublisher.NewEonKeyPublisher(
		kpr.dbpool,
		kpr.gnosisEthClient,
		kpr.config.Gnosis.Contracts.KeyperSetManager,
		kpr.config.Gnosis.Node.PrivateKey.Key,
	)
	if err != nil {
		return errors.Wrap(err, "failed to initialize eon key publisher")
	}

	validatorUpdated, err := synchandler.NewValidatorUpdated(
		kpr.dbpool,
		kpr.gnosisEthClient,
		kpr.beaconAPIClient,
		kpr.config.Gnosis.Contracts.ValidatorRegistry,
		chainID.Uint64(),
	)
	if err != nil {
		return err
	}
	sequencerTxSubmitted, err := synchandler.NewSequencerTransactionSubmitted(
		kpr.dbpool,
		kpr.config.Gnosis.Contracts.Sequencer,
	)
	if err != nil {
		return err
	}
	err = kpr.initKeyperCore(
		messaging,
		[]syncer.ChainUpdateHandler{
			// trigger possible decryption on every new block header from Gnosis chain
			synchandler.NewDecryptOnChainUpdate(kpr.maybeDecryptOnNewHeader),
		},
		[]syncer.ContractEventHandler{
			// process the SequencerTransactionSubmitted events from Gnosis chain
			sequencerTxSubmitted,
			// process the ValidatorUpdated events from Gnosis chain
			validatorUpdated,
		},
	)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
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
	runner.Go(func() error { return kpr.processInputs(ctx) })
	return runner.StartService(kpr.core, kpr.slotTicker, kpr.eonKeyPublisher)
}

func (kpr *Keyper) initKeyperCore(
	messaging *p2p.P2PMessaging,
	chainUpdateHandler []syncer.ChainUpdateHandler,
	eventHandler []syncer.ContractEventHandler,
) error {
	opts := []keyper.Option{
		// re-use the gnosis client to receive chain updates
		keyper.WithBlockSyncClient(kpr.gnosisEthClient),
		keyper.WithDBPool(kpr.dbpool),
		// P2P messaging - we use a middleware to inject some special
		// functionality
		keyper.WithMessaging(
			NewMessagingMiddleware(messaging, kpr.dbpool, kpr.config),
		),

		// don't broadcast generated eon public keys to the P2P network,
		keyper.NoBroadcastEonPublicKey(),
		// but instead push them on the internal channel,
		// where it gets written to a contract onchain
		keyper.WithEonPublicKeyHandler(kpr.sendNewEonPubkeyToChannel),

		// This will only start syncing blockchain events
		// from a specific block on, and only when we never synced before.
		// Otherwise, it will pick up syncing where we last stopped.
		// Beware that this has to be a block at or before where the keyper-set was
		// added (not started) onchain.
		// Otherwise the node will miss the event and never know it is
		// part of the keyper-set.
		keyper.WithSyncStartBlockNumber(*new(big.Int).SetUint64(kpr.config.Gnosis.SyncStartBlockNumber)),
	}
	for _, h := range chainUpdateHandler {
		opts = append(opts, keyper.WithChainUpdateHandler(h))
	}
	for _, h := range eventHandler {
		opts = append(opts, keyper.WithContractEventHandler(h))
	}
	var err error
	kpr.core, err = keyper.New(
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
			// The keyper core needs the implementation addresses of the core
			// contracts on the specific chain.
			ContractAddresses: kprconfig.ContractAddresses{
				KeyperSetManager: kpr.config.Gnosis.Contracts.KeyperSetManager,
			},
		},
		// send events to this channel to trigger decryption in the keyper core
		kpr.decryptionTriggerChannel,
		opts...,
	)
	return err
}

func (kpr *Keyper) maybeDecryptOnNewHeader(ctx context.Context, header *types.Header) error {
	slot := medley.BlockTimestampToSlot(
		header.Time,
		kpr.config.Gnosis.GenesisSlotTimestamp,
		kpr.config.Gnosis.SecondsPerSlot,
	)
	return kpr.maybeDecryptOnNewSlot(ctx, slot+1)
}

func (kpr *Keyper) processInputs(ctx context.Context) error {
	for {
		select {
		case key := <-kpr.newEonPublicKeys:
			kpr.eonKeyPublisher.Publish(key)
		case slot := <-kpr.slotTicker.C:
			logger := log.Logger.With().Uint64("slot-number", slot.Number).Time("slot-start", slot.Start()).Logger()
			logger.Debug().Msg("slot ticker fired, try decrypting")
			if err := kpr.maybeDecryptOnNewSlot(ctx, slot.Number); err != nil {
				// TODO: Check if it's safe to drop those events. If not, we should store the
				// ones that remain on the channel in the db and process them when we restart.
				if errors.Is(err, errs.ErrCritical) {
					return err
				}
				logger.Error().Err(err).Msg("error trying to decrypt on new slot")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (kpr *Keyper) sendNewEonPubkeyToChannel(_ context.Context, key keyper.EonPublicKey) error {
	kpr.newEonPublicKeys <- key
	return nil
}
