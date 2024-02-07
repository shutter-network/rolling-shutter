package gnosis

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

var ErrParseKeyperSet = errors.New("can't parse KeyperSet")

type Keyper struct {
	core            *keyper.KeyperCore
	config          *Config
	dbpool          *pgxpool.Pool
	chainSyncClient *chainsync.Client

	sequencerSyncer *SequencerSyncer
}

func New(c *Config) *Keyper {
	return &Keyper{
		config: c,
	}
}

func (kpr *Keyper) Start(ctx context.Context, runner service.Runner) error {
	var err error

	decrTrigChan := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])
	runner.Defer(func() { close(decrTrigChan) })

	kpr.dbpool, err = db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	kpr.core, err = keyper.New(
		&kprconfig.Config{
			InstanceID:        kpr.config.InstanceID,
			DatabaseURL:       kpr.config.DatabaseURL,
			HTTPEnabled:       kpr.config.HTTPEnabled,
			HTTPListenAddress: kpr.config.HTTPListenAddress,
			P2P:               kpr.config.P2P,
			Ethereum:          kpr.config.Gnosis,
			Shuttermint:       kpr.config.Shuttermint,
			Metrics:           kpr.config.Metrics,
		},
		decrTrigChan,
		keyper.WithDBPool(kpr.dbpool),
		keyper.NoBroadcastEonPublicKey(),
		keyper.WithEonPublicKeyHandler(kpr.newEonPublicKey),
	)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
	}

	triggerer := NewDecryptionTriggerer(kpr.config.Gnosis, decrTrigChan)

	kpr.chainSyncClient, err = chainsync.NewClient(
		ctx,
		chainsync.WithClientURL(kpr.config.Gnosis.EthereumURL),
		chainsync.WithKeyperSetManager(kpr.config.GnosisContracts.KeyperSetManager),
		chainsync.WithKeyBroadcastContract(kpr.config.GnosisContracts.KeyBroadcastContract),
		chainsync.WithSyncNewBlock(kpr.newBlock),
		chainsync.WithSyncNewKeyperSet(kpr.newKeyperSet),
		chainsync.WithPrivateKey(kpr.config.Gnosis.PrivateKey.Key),
	)
	if err != nil {
		return err
	}

	err = kpr.initSequencerSyncer(ctx)
	if err != nil {
		return err
	}

	return runner.StartService(kpr.core, triggerer, kpr.chainSyncClient)
}

// initSequencerSycer initializes the sequencer syncer if the keyper is known to be a member of a
// keyper set. Otherwise, the syncer will only be initialied once such a keyper set is observed to
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
			Str("contract-address", kpr.config.GnosisContracts.KeyperSetManager.Hex()).
			Msg("initializing sequencer syncer")
		client, err := ethclient.DialContext(ctx, kpr.config.Gnosis.ContractsURL)
		if err != nil {
			return err
		}
		contract, err := sequencerBindings.NewSequencer(kpr.config.GnosisContracts.Sequencer, client)
		if err != nil {
			return err
		}
		kpr.sequencerSyncer = &SequencerSyncer{
			Contract: contract,
			DBPool:   kpr.dbpool,
			StartEon: eon,
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

func (kpr *Keyper) newBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	if kpr.sequencerSyncer != nil {
		if err := kpr.sequencerSyncer.Sync(ctx, ev.Number.Uint64()); err != nil {
			return err
		}
	}
	return nil
}

func (kpr *Keyper) newKeyperSet(ctx context.Context, ev *syncevent.KeyperSet) error {
	isMember := false
	for _, m := range ev.Members {
		if m.Cmp(kpr.config.GetAddress()) == 0 {
			isMember = true
			break
		}
	}
	log.Info().
		Uint64("activation-block", ev.ActivationBlock).
		Uint64("eon", ev.Eon).
		Bool("is-member", isMember).
		Msg("new keyper set added")

	if isMember {
		if err := kpr.ensureSequencerSyncing(ctx, ev.Eon); err != nil {
			return err
		}
	}

	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		obskeyperdb := obskeyper.New(tx)

		keyperConfigIndex, err := medley.Uint64ToInt64Safe(ev.Eon)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		activationBlockNumber, err := medley.Uint64ToInt64Safe(ev.ActivationBlock)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		threshold, err := medley.Uint64ToInt64Safe(ev.Threshold)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}

		return obskeyperdb.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
			KeyperConfigIndex:     keyperConfigIndex,
			ActivationBlockNumber: activationBlockNumber,
			Keypers:               shdb.EncodeAddresses(ev.Members),
			Threshold:             int32(threshold),
		})
	})
}

func (kpr *Keyper) newEonPublicKey(_ context.Context, pubKey keyper.EonPublicKey) error {
	log.Info().
		Uint64("eon", pubKey.Eon).
		Uint64("activation-block", pubKey.ActivationBlock).
		Msg("new eon pk")
	return nil
}
