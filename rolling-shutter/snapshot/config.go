package snapshot

import (
	"io"

	"github.com/multiformats/go-multiaddr"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/address"
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
	c.Ethereum = configuration.NewEthnodeConfig()
	c.Metrics = metricsserver.NewConfig()
}

type Config struct {
	InstanceID     uint64 `shconfig:",required"`
	DatabaseURL    string `shconfig:",required"`
	SnapshotHubURL string `shconfig:",required"`

	JSONRPCHost string
	JSONRPCPort uint16

	P2P      *p2p.Config
	Ethereum *configuration.EthnodeConfig
	Metrics  *metricsserver.MetricsConfig
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Name() string {
	return "snapshot"
}

func (c *Config) SetDefaultValues() error {
	// overwrite the child config's usual default value
	c.Ethereum.EthereumURL = "http://[::1]:8545/"
	c.JSONRPCHost = ""
	c.JSONRPCPort = 8754
	c.Metrics.Enabled = false
	c.Metrics.Host = "127.0.0.1"
	c.Metrics.Port = 9191
	return nil
}

func (c *Config) SetExampleValues() error {
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}
	listenAddr, err := multiaddr.NewMultiaddr("/ip6/::1/tcp/2000")
	if err != nil {
		return err
	}
	// overwrite the child config's usual example value
	c.P2P.ListenAddresses = []*address.P2PAddress{{Multiaddr: listenAddr}}
	c.InstanceID = 42
	c.DatabaseURL = "postgres://pguser:pgpassword@localhost:5432/shutter_snapshot"
	return nil
}

func (c Config) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}
