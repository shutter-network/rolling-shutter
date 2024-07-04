package gnosisaccessnode

import (
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type Config struct {
	InstanceID uint64

	P2P     *p2p.Config
	Metrics *metricsserver.MetricsConfig

	MaxNumKeysPerMessage uint64
}

func (c *Config) Init() {
	c.P2P = p2p.NewConfig()
	c.Metrics = metricsserver.NewConfig()
}

func (c *Config) Name() string {
	return "gnosis-access-node"
}

func (c *Config) Validate() error {
	if err := c.P2P.Validate(); err != nil {
		return err
	}
	if err := c.Metrics.Validate(); err != nil {
		return err
	}
	return nil
}

func (c *Config) SetDefaultValues() error {
	return nil
}

func (c *Config) SetExampleValues() error { //nolint:unparam
	c.InstanceID = 1000
	c.MaxNumKeysPerMessage = 500
	return nil
}

func (c *Config) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}
