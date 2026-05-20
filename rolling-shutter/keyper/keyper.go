package keyper

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/keypermetrics"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprapi"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/channel"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type KeyperCore struct {
	trigger <-chan *broker.Event[*epochkghandler.DecryptionTrigger]
	opts    *options
	config  *kprconfig.Config

	dbpool          *pgxpool.Pool
	messaging       p2p.Messaging
	blockSyncClient *ethclient.Client

	metricsServer *metricsserver.MetricsServer
}

func New(
	config *kprconfig.Config,
	trigger <-chan *broker.Event[*epochkghandler.DecryptionTrigger],
	options ...Option,
) (*KeyperCore, error) {
	opts := newDefaultOptions()
	for _, option := range options {
		err := option(opts)
		if err != nil {
			return nil, err
		}
	}
	sender := opts.messaging
	if sender == nil {
		var err error
		sender, err = p2p.New(config.P2P)
		if err != nil {
			return nil, err
		}
	}
	return &KeyperCore{config: config, trigger: trigger, messaging: sender, opts: opts}, nil
}

// LinkConfigToDB ensures that we use a database compatible with the given config. On first use
// it stores the config's ethereum address into the database. On subsequent uses it compares the
// stored value and raises an error if it doesn't match.
func LinkConfigToDB(ctx context.Context, config *kprconfig.Config, dbpool *pgxpool.Pool) error {
	const addressKey = "ethereum address"
	cfgAddress := config.GetAddress().String()
	queries := db.New(dbpool)
	dbAddr, err := queries.GetMeta(ctx, addressKey)
	if err == pgx.ErrNoRows {
		return queries.InsertMeta(ctx, db.InsertMetaParams{
			Key:   addressKey,
			Value: cfgAddress,
		})
	} else if err != nil {
		return err
	}

	if dbAddr != cfgAddress {
		return errors.Errorf(
			"database linked to wrong address %s, config address is %s",
			dbAddr, cfgAddress)
	}
	return nil
}

func (kpr *KeyperCore) initOptions(ctx context.Context, runner service.Runner) error {
	if kpr.opts.dbpool == nil {
		// connect, but don't validate any database version.
		// If that is desired, it should be done in the keyper-implementation
		var err error
		kpr.dbpool, err = db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
		if err != nil {
			return err
		}
		runner.Defer(kpr.dbpool.Close)
	} else {
		kpr.dbpool = kpr.opts.dbpool
	}
	if kpr.opts.blockSyncClient == nil {
		var err error
		kpr.blockSyncClient, err = ethclient.DialContext(ctx, kpr.config.Ethereum.EthereumURL)
		if err != nil {
			return err
		}
	} else {
		kpr.blockSyncClient = kpr.opts.blockSyncClient
	}
	return nil
}

func (kpr *KeyperCore) Start(ctx context.Context, runner service.Runner) error {
	config := kpr.config
	err := kpr.initOptions(ctx, runner)
	if err != nil {
		return err
	}

	err = kpr.dbpool.BeginFunc(db.WrapContext(ctx, database.Definition.Validate))
	if err != nil {
		return err
	}
	err = LinkConfigToDB(ctx, config, kpr.dbpool)
	if err != nil {
		return err
	}

	if kpr.config.Metrics.Enabled {
		keypermetrics.InitMetrics(kpr.dbpool, *kpr.config)
		epochkghandler.InitMetrics()
		deployment.InitMetrics()
		kpr.metricsServer = metricsserver.New(kpr.config.Metrics)
	}

	kpr.messaging.AddMessageHandler(
		epochkghandler.NewDecryptionKeyHandler(kpr.config, kpr.dbpool),
		epochkghandler.NewDecryptionKeyShareHandler(kpr.config, kpr.dbpool),
		// this is purely used to subscribe to the public key topic for broadcast
		epochkghandler.NewEonPublicKeyHandler(kpr.config, kpr.dbpool),
	)
	kpr.messaging.AddMessageHandler(kpr.opts.messageHandler...)
	return runner.StartService(kpr.getServices()...)
}

func (kpr *KeyperCore) getServices() []service.Service {
	services := []service.Service{
		kpr.messaging,
	}
	keyTrigger := kpr.trigger
	if kpr.config.HTTPEnabled {
		httpServer := kprapi.NewHTTPService(kpr.dbpool, kpr.config, kpr.messaging)
		services = append(services, httpServer)
		// combine two sources of decryption triggers
		// and spawn the fan-in routine
		apiDecrTrig := httpServer.GetDecryptionTriggerChannel()
		fanIn := channel.NewFanInService(kpr.trigger, apiDecrTrig)
		services = append(services, fanIn)
		keyTrigger = fanIn.C
	}
	keyShareHandler := &epochkghandler.KeyShareHandler{
		InstanceID:           kpr.config.GetInstanceID(),
		KeyperAddress:        kpr.config.GetAddress(),
		MaxNumKeysPerMessage: kpr.config.GetMaxNumKeysPerMessage(),
		DBPool:               kpr.dbpool,
		Messaging:            kpr.messaging,
		Trigger:              keyTrigger,
	}
	services = append(services, keyShareHandler)
	if kpr.config.Metrics.Enabled {
		services = append(services, kpr.metricsServer)
	}
	return services
}
