package gnosis

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

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

	return runner.StartService(kpr.core, kpr.chainSyncClient)
}

func (kpr *Keyper) newBlock(_ context.Context, ev *syncevent.LatestBlock) error {
	log.Info().
		Uint64("number", ev.Number.Uint64()).
		Str("hash", ev.BlockHash.Hex()).
		Msg("new latest block")
	return nil
}

func (kpr *Keyper) newKeyperSet(ctx context.Context, ev *syncevent.KeyperSet) error {
	log.Info().
		Uint64("activation-block", ev.ActivationBlock).
		Uint64("eon", ev.Eon).
		Msg("new keyper set added")
	fmt.Printf("%+v\n", ev.Members)

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
