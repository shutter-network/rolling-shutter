package gnosis

import (
	"context"
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
	config          *config.Config
	dbpool          *pgxpool.Pool
	beaconAPIClient *beaconapiclient.Client

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

	messageSender, err := p2p.New(kpr.config.P2P)
	if err != nil {
		return errors.Wrap(err, "failed to initialize p2p messaging")
	}
	messageSender.AddMessageHandler(&DecryptionKeySharesHandler{kpr.dbpool})
	messageSender.AddMessageHandler(&DecryptionKeysHandler{kpr.dbpool})
	messagingMiddleware := NewMessagingMiddleware(messageSender, kpr.dbpool, kpr.config)

	kpr.core, err = InitializeKeyperCore(ctx, kpr, messagingMiddleware)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
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

func InitializeKeyperCore(ctx context.Context, kpr *Keyper, messagingMiddleware *MessagingMiddleware) (*keyper.KeyperCore, error) {
	ethClient, err := ethclient.DialContext(ctx, kpr.config.Gnosis.Node.EthereumURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial ethereum node")
	}
	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chain ID")
	}

	validatorUpdated, err := synchandler.NewValidatorUpdated(
		kpr.dbpool,
		ethClient,
		kpr.beaconAPIClient,
		kpr.config.Gnosis.Contracts.ValidatorRegistry,
		chainID.Uint64(),
	)
	if err != nil {
		return nil, err
	}
	sequencerTxSubmitted, err := synchandler.NewSequencerTransactionSubmitted(
		kpr.dbpool,
		kpr.config.Gnosis.Contracts.Sequencer,
	)
	if err != nil {
		return nil, err
	}
	decryptOnChainUpdate := synchandler.NewDecryptOnChainUpdateHandler(kpr.newBlockHandlerFunc)
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
		keyper.WithContractAddresses(
			keyper.ContractAddresses{
				KeyperSetManager: kpr.config.Gnosis.Contracts.KeyperSetManager,
			}),
		keyper.WithDBPool(kpr.dbpool),
		keyper.WithMessaging(messagingMiddleware),

		keyper.NoBroadcastEonPublicKey(),
		keyper.WithEonPublicKeyHandler(kpr.sendNewEonPubkeyToChannel),

		keyper.WithChainUpdateHandler(decryptOnChainUpdate),
		keyper.WithContractEventHandler(sequencerTxSubmitted),
		keyper.WithContractEventHandler(validatorUpdated),
	)
	return core, err
}

// initSequencerSyncer initializes the sequencer syncer if the keyper is known to be a member of a
// keyper set. Otherwise, the syncer will only be initialized once such a keyper set is observed to
// be added, as only then we will know which eon(s) we are responsible for.
func (kpr *Keyper) initSequencerSyncer(ctx context.Context) error {
	// FIXME: is the described "once" behavior still the same with the handler implementation?
	return nil
}

func (kpr *Keyper) newBlockHandlerFunc(ctx context.Context, header *types.Header) error {
	slot := medley.BlockTimestampToSlot(
		header.Time,
		kpr.config.Gnosis.GenesisSlotTimestamp,
		kpr.config.Gnosis.SecondsPerSlot,
	)
	return kpr.maybeTriggerDecryption(ctx, slot+1)
}

func (kpr *Keyper) processInputs(ctx context.Context) error {
	var err error
	for {
		select {
		case key := <-kpr.newEonPublicKeys:
			kpr.eonKeyPublisher.Publish(key)
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

func (kpr *Keyper) sendNewEonPubkeyToChannel(_ context.Context, key keyper.EonPublicKey) error {
	kpr.newEonPublicKeys <- key
	return nil
}
