package optimism

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

var ErrParseKeyperSet = errors.New("can't parse KeyperSet")

type Keyper struct {
	core     *keyper.KeyperCore
	l2Client *chainsync.Client
	dbpool   *pgxpool.Pool
	config   *config.Config

	trigger chan<- *broker.Event[*epochkghandler.DecryptionTrigger]
}

func New(c *config.Config) (*Keyper, error) { //nolint:unparam
	return &Keyper{
		config: c,
	}, nil
}

func (kpr *Keyper) Start(ctx context.Context, runner service.Runner) error {
	var err error

	dbpool, err := db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	kpr.dbpool = dbpool

	trigger := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])
	kpr.trigger = trigger

	ethConfig := configuration.NewEthnodeConfig()
	ethConfig.EthereumURL = kpr.config.Optimism.JSONRPCURL
	ethConfig.PrivateKey = kpr.config.Optimism.PrivateKey
	kpr.core, err = keyper.New(
		&kprconfig.Config{
			InstanceID:        kpr.config.InstanceID,
			DatabaseURL:       kpr.config.DatabaseURL,
			HTTPEnabled:       kpr.config.HTTPEnabled,
			HTTPListenAddress: kpr.config.HTTPListenAddress,
			P2P:               kpr.config.P2P,
			Ethereum:          ethConfig,
			Shuttermint:       kpr.config.Shuttermint,
			Metrics:           kpr.config.Metrics,
		},
		trigger,
		keyper.WithDBPool(dbpool),
		keyper.NoBroadcastEonPublicKey(),
		keyper.WithEonPublicKeyHandler(kpr.newEonPublicKey),
	)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
	}
	// TODO: wrap the logger and pass in
	kpr.l2Client, err = chainsync.NewClient(
		ctx,
		chainsync.WithClientURL(kpr.config.Optimism.JSONRPCURL),
		chainsync.WithSyncNewBlock(kpr.newBlock),
		chainsync.WithSyncNewKeyperSet(kpr.newKeyperSet),
		chainsync.WithPrivateKey(kpr.config.Optimism.PrivateKey.Key),
	)
	if err != nil {
		return err
	}
	return runner.StartService(kpr.core, kpr.l2Client)
}

func (kpr *Keyper) newBlock(_ context.Context, ev *syncevent.LatestBlock) error {
	log.Info().
		Int64("number", ev.Number.Int64()).
		Str("hash", ev.BlockHash.Hex()).
		Msg("new latest unsafe head on L2")

	// TODO: sanity checks

	latestBlockNumber := ev.Number.Uint64()
	idPreimage := identitypreimage.Uint64ToIdentityPreimage(latestBlockNumber + 1)
	trig := &epochkghandler.DecryptionTrigger{
		BlockNumber:       latestBlockNumber + 1,
		IdentityPreimages: []identitypreimage.IdentityPreimage{idPreimage},
	}

	// TODO: what to do if this blocks?
	kpr.trigger <- broker.NewEvent(trig)
	return nil
}

func (kpr *Keyper) newKeyperSet(ctx context.Context, ev *syncevent.KeyperSet) error {
	log.Info().
		Uint64("activation-block", ev.ActivationBlock).
		Uint64("index", ev.Index).
		Msg("new keyper set added")
	return kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		obskeyperdb := obskeyper.New(tx)

		keyperConfigIndex, err := medley.Uint64ToInt64Safe(ev.Index)
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
		// XXX: does this work when the memberset is empty?
		return obskeyperdb.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
			KeyperConfigIndex:     keyperConfigIndex,
			ActivationBlockNumber: activationBlockNumber,
			Keypers:               shdb.EncodeAddresses(ev.Members),
			Threshold:             int32(threshold),
		})
	})
}

func (kpr *Keyper) newEonPublicKey(ctx context.Context, pubKey keyper.EonPublicKey) error {
	log.Info().
		Uint64("eon", pubKey.Eon).
		Uint64("activation-block", pubKey.ActivationBlock).
		Bytes("pub-key", pubKey.PublicKey).
		Msg("new eon pk")
	// Currently all keypers call this and race to call this function first.
	// For now this is fine, but a keyper should only send a transaction if
	// the key is not set yet.
	// Best would be a coordinatated leader election who will broadcast the key.
	// FIXME: the syncer receives an empty key byte.
	// Is this already
	tx, err := kpr.l2Client.BroadcastEonKey(ctx, pubKey.Eon, pubKey.PublicKey)
	if err != nil {
		log.Error().Err(err).Msg("error broadcasting eon public key")
		return errors.Wrap(err, "error broadcasting eon public-key")
	}
	log.Info().
		Str("hash", tx.Hash().Hex()).
		Msg("sent eon pubkey transaction")

	receipt, err := bind.WaitMined(ctx, kpr.l2Client, tx)
	if err != nil {
		log.Error().Err(err).Msg("error waiting for pubkey tx mined")
		return err
	}
	// NOCHECKIN: log the JSON receipt or only specific fields
	log.Info().
		Interface("receipt", receipt).
		Msg("eon pubkey transaction mined")
	// TODO:
	// wait / confirm of tx, otherwise resend
	return nil
}
