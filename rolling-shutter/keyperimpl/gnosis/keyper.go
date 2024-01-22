package gnosis

import (
	"context"

	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/db"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type Keyper struct {
	core   *keyper.KeyperCore
	config *Config
}

func New(c *Config) *Keyper {
	return &Keyper{
		config: c,
	}
}

func (kpr *Keyper) Start(ctx context.Context, runner service.Runner) error {
	decrTrigChan := make(chan *broker.Event[*epochkghandler.DecryptionTrigger])
	runner.Defer(func() { close(decrTrigChan) })

	dbpool, err := db.Connect(ctx, runner, kpr.config.DatabaseURL, database.Definition.Name())
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
		keyper.WithDBPool(dbpool),
	)
	if err != nil {
		return errors.Wrap(err, "can't instantiate keyper core")
	}

	return runner.StartService(kpr.core)
}
