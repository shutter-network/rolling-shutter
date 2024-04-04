package gnosis

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gethLog "github.com/ethereum/go-ethereum/log"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	"golang.org/x/exp/slog"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/slotticker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

var ErrParseKeyperSet = errors.New("cannot parse KeyperSet")

type Keyper struct {
	core            *keyper.KeyperCore
	config          *Config
	dbpool          *pgxpool.Pool
	chainSyncClient *chainsync.Client
	sequencerSyncer *SequencerSyncer

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
	runner.Defer(func() { close(kpr.newBlocks) })
	runner.Defer(func() { close(kpr.newKeyperSets) })
	runner.Defer(func() { close(kpr.newEonPublicKeys) })
	runner.Defer(func() { close(kpr.decryptionTriggerChannel) })

	kpr.slotTicker = slotticker.NewSlotTicker(
		time.Duration(kpr.config.Gnosis.SecondsPerSlot*uint64(time.Second)),
		time.Unix(int64(kpr.config.Gnosis.GenesisSlotTimestamp), 0),
	)

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

	kpr.core, err = keyper.New(
		&kprconfig.Config{
			InstanceID:        kpr.config.InstanceID,
			DatabaseURL:       kpr.config.DatabaseURL,
			HTTPEnabled:       kpr.config.HTTPEnabled,
			HTTPListenAddress: kpr.config.HTTPListenAddress,
			P2P:               kpr.config.P2P,
			Ethereum:          kpr.config.Gnosis.Node,
			Shuttermint:       kpr.config.Shuttermint,
			Metrics:           kpr.config.Metrics,
		},
		kpr.decryptionTriggerChannel,
		keyper.WithDBPool(kpr.dbpool),
		keyper.NoBroadcastEonPublicKey(),
		keyper.WithEonPublicKeyHandler(kpr.channelNewEonPublicKey),
		keyper.WithMessaging(messagingMiddleware),
	)
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

	err = kpr.initSequencerSyncer(ctx)
	if err != nil {
		return err
	}

	runner.Go(func() error { return kpr.processInputs(ctx) })
	return runner.StartService(kpr.core, kpr.chainSyncClient, kpr.slotTicker)
}

// initSequencerSycer initializes the sequencer syncer if the keyper is known to be a member of a
// keyper set. Otherwise, the syncer will only be initialized once such a keyper set is observed to
// be added, as only then we will know which eon(s) we are responsible for.
func (kpr *Keyper) initSequencerSyncer(ctx context.Context) error {
	obskeyperdb := obskeyper.New(kpr.dbpool)
	keyperSets, err := obskeyperdb.GetKeyperSets(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to query keyper sets from db")
	}

	keyperSetFound := false
	minEon := uint64(0)
	for _, keyperSet := range keyperSets {
		for _, m := range keyperSet.Keypers {
			mAddress := common.HexToAddress(m)
			if mAddress.Cmp(kpr.config.GetAddress()) == 0 {
				keyperSetFound = true
				if minEon > uint64(keyperSet.KeyperConfigIndex) {
					minEon = uint64(keyperSet.KeyperConfigIndex)
				}
				break
			}
		}
	}

	if keyperSetFound {
		err := kpr.ensureSequencerSyncing(ctx, minEon)
		if err != nil {
			return err
		}
	}
	return nil
}

func (kpr *Keyper) ensureSequencerSyncing(ctx context.Context, eon uint64) error {
	if kpr.sequencerSyncer == nil {
		log.Info().
			Uint64("eon", eon).
			Str("contract-address", kpr.config.Gnosis.Contracts.KeyperSetManager.Hex()).
			Msg("initializing sequencer syncer")
		client, err := ethclient.DialContext(ctx, kpr.config.Gnosis.Node.ContractsURL)
		if err != nil {
			return err
		}
		contract, err := sequencerBindings.NewSequencer(kpr.config.Gnosis.Contracts.Sequencer, client)
		if err != nil {
			return err
		}
		kpr.sequencerSyncer = &SequencerSyncer{
			Contract:             contract,
			DBPool:               kpr.dbpool,
			StartEon:             eon,
			GenesisSlotTimestamp: kpr.config.Gnosis.GenesisSlotTimestamp,
			SecondsPerSlot:       kpr.config.Gnosis.SecondsPerSlot,
		}

		// TODO: perform an initial sync without blocking and/or set start block
	}

	if eon < kpr.sequencerSyncer.StartEon {
		log.Info().
			Uint64("old-start-eon", kpr.sequencerSyncer.StartEon).
			Uint64("new-start-eon", eon).
			Msg("decreasing sequencer syncing start eon")
		kpr.sequencerSyncer.StartEon = eon
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
