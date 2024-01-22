package gnosis

import (
	"io"

	"github.com/ethereum/go-ethereum/common"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

var _ configuration.Config = &Config{}

func NewConfig() *Config {
	c := &Config{}
	c.Init()
	return c
}

func (c *Config) Init() {
	c.P2P = p2p.NewConfig()
	c.Gnosis = configuration.NewEthnodeConfig()
	c.Shuttermint = kprconfig.NewShuttermintConfig()
	c.Metrics = metricsserver.NewConfig()
}

type Config struct {
	InstanceID  uint64 `shconfig:",required"`
	DatabaseURL string `shconfig:",required" comment:"If it's empty, we use the standard PG_ environment variables"`

	HTTPEnabled       bool
	HTTPListenAddress string

	P2P         *p2p.Config
	Gnosis      *configuration.EthnodeConfig
	Shuttermint *kprconfig.ShuttermintConfig
	Metrics     *metricsserver.MetricsConfig
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Name() string {
	return "gnosiskeyper"
}

func (c *Config) SetDefaultValues() error {
	c.HTTPEnabled = false
	c.HTTPListenAddress = ":3000"
	return nil
}

func (c *Config) SetExampleValues() error {
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}
	c.InstanceID = 42
	c.DatabaseURL = "postgres://pguser:pgpassword@localhost:5432/shutter"
	return nil
}

func (c Config) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}

func (c *Config) GetAddress() common.Address {
	return c.Gnosis.PrivateKey.EthereumAddress()
}
