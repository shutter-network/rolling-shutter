package mocksequencer

import (
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
)

var _ configuration.Config = &Config{}

func NewConfig() *Config {
	return &Config{}
}

type Config struct {
	ChainID     uint64 `shconfig:",required"`
	EthereumURL string

	HTTPListenAddress string

	MaxBlockDeviation    uint64
	EthereumPollInterval uint64

	Admin bool
	Debug bool
}

func (c *Config) Init() {
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Name() string {
	return "mocksequencer"
}

func (c *Config) SetDefaultValues() error {
	c.HTTPListenAddress = "localhost:8555"
	c.EthereumURL = "http://localhost:8545"
	c.MaxBlockDeviation = 5
	c.EthereumPollInterval = 1
	c.Admin = true
	c.Debug = true
	return nil
}

func (c *Config) SetExampleValues() error {
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}
	c.ChainID = 42
	return nil
}

func (c Config) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}
