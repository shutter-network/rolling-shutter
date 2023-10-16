package mocksequencer

import (
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
)

var _ configuration.Config = &Config{}

func NewConfig() *Config {
	return &Config{}
}

type Config struct {
	DatabaseURL   string `shconfig:",required" comment:"If it's empty, we use the standard PG_ environment variables"`
	DeploymentDir string `                     comment:"Contract source directory"`
	L2BackendURL  *url.URL

	HTTPListenAddress string

	MaxBlockDeviation    uint64
	EthereumPollInterval uint64

	Admin bool
	Debug bool

	// P2P *p2p.Config
}

func (c *Config) Init() {
	c.L2BackendURL = &url.URL{}
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Name() string {
	return "mocksequencer"
}

func (c *Config) SetDefaultValues() error {
	c.HTTPListenAddress = "localhost:8555"
	c.DeploymentDir = "./deployments/localhost/"

	err := c.L2BackendURL.UnmarshalText([]byte("http://127.0.0.1:8545/"))
	if err != nil {
		return err
	}
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
	// c.ChainID = 42
	c.DatabaseURL = "postgres://pguser:pgpassword@localhost:5432/shutter"
	return nil
}

func (c Config) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}
