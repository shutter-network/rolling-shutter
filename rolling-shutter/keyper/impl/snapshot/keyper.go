package snapshot

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract/deployment"
	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/db/chainobsdb/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/eventsyncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type Keyper struct {
	core      *keyper.KeyperCore
	contracts *deployment.Contracts
	dbpool    *pgxpool.Pool
	config    *Config

	trigger chan<- *broker.Event[*epochkghandler.DecryptionTrigger]
}

func New(c *Config) (*Keyper, error) {
	decrTrigChan := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])
	coreConfig := &kprconfig.Config{
		InstanceID:         c.InstanceID,
		DatabaseURL:        c.DatabaseURL,
		HTTPEnabled:        c.HTTPEnabled,
		HTTPListenAddress:  c.HTTPListenAddress,
		P2P:                c.P2P,
		EthereumPrivateKey: c.Ethereum.PrivateKey,
		EthereumURL:        c.Ethereum.EthereumURL,
		Shuttermint:        c.Shuttermint,
		Metrics:            c.Metrics,
	}
	core, err := keyper.New(
		coreConfig,
		decrTrigChan,
	)
	if err != nil {
		return nil, errors.Wrap(err, "can't instantiate keyper core")
	}
	return &Keyper{
		config:  c,
		core:    core,
		trigger: decrTrigChan,
	}, nil
}

func (kpr *Keyper) Start(ctx context.Context, runner service.Runner) error {
	// FIXME this is a different dbpool, since the core-keyper instantiates
	// another one. Maybe this is fine, maybe we can reuse...
	dbpool, err := pgxpool.Connect(ctx, kpr.config.DatabaseURL)
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

	kpr.contracts = contracts
	kpr.dbpool = dbpool

	runner.StartService(kpr.core)

	services := []service.Service{
		service.ServiceFn{Fn: kpr.handleContractEvents},
	}
	// TODO different services
	return runner.StartService(services...)
}

func (kpr *Keyper) handleContractEvents(ctx context.Context) error {
	// XXX this might stay part of the core?
	kprHandler := &obskeyper.Handler{
		KeyperContract: kpr.contracts.Keypers,
	}
	events := map[*eventsyncer.EventType]chainobserver.EventHandlerFunc{
		kpr.contracts.KeypersConfigsListNewConfig: chainobserver.MakeHandler(kprHandler.HandleKeypersConfigsListNewConfigEvent),
	}
	return chainobserver.New(kpr.contracts, kpr.dbpool).Observe(ctx, events)
}
