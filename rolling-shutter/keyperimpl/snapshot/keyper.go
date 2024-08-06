package snapshot

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/snapshot/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type Keyper struct {
	core   *keyper.KeyperCore
	config *Config

	trigger chan<- *broker.Event[*epochkghandler.DecryptionTrigger]
}

func New(c *Config) *Keyper {
	return &Keyper{
		config: c,
	}
}

func (kpr *Keyper) Start(ctx context.Context, runner service.Runner) error {
	var err error

	decrTrigChan := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])
	kpr.trigger = decrTrigChan
	runner.Defer(func() { close(decrTrigChan) })

	dbpool, err := db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	contractsClient, err := ethclient.Dial(kpr.config.Ethereum.ContractsURL)
	if err != nil {
		return err
	}
	contracts, err := deployment.NewContracts(contractsClient, kpr.config.Ethereum.DeploymentDir)
	if err != nil {
		return err
	}
	listenEvents := []*eventsyncer.EventType{
		contracts.CollatorConfigsListNewConfig,
		contracts.KeypersConfigsListNewConfig,
	}
	chainobs := chainobserver.New(contractsClient, dbpool)
	for _, ev := range listenEvents {
		if err := chainobs.AddListenEvent(ev); err != nil {
			return err
		}
	}
	kpr.core, err = keyper.New(
		&kprconfig.Config{
			InstanceID:           kpr.config.InstanceID,
			DatabaseURL:          kpr.config.DatabaseURL,
			HTTPEnabled:          kpr.config.HTTPEnabled,
			HTTPListenAddress:    kpr.config.HTTPListenAddress,
			P2P:                  kpr.config.P2P,
			Ethereum:             kpr.config.Ethereum,
			Shuttermint:          kpr.config.Shuttermint,
			Metrics:              kpr.config.Metrics,
			MaxNumKeysPerMessage: kpr.config.MaxNumKeysPerMessage,
		},
		decrTrigChan,
		keyper.WithMessageHandler(NewDecryptionTriggerHandler(*kpr.config, dbpool, kpr.trigger)),
	)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
	}

	return runner.StartService(kpr.core, chainobs)
}
