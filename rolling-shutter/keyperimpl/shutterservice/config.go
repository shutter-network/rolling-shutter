package shutterservice

import (
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

func NewConfig() *Config {
	c := &Config{}
	c.Init()
	return c
}

func (c *Config) Init() {
	c.P2P = p2p.NewConfig()
	c.Shuttermint = kprconfig.NewShuttermintConfig()
	c.Metrics = metricsserver.NewConfig()
}

type Config struct {
	InstanceID  uint64 `shconfig:",required"`
	DatabaseURL string `shconfig:",required" comment:"If it's empty, we use the standard PG_ environment variables"`

	HTTPEnabled       bool
	HTTPListenAddress string

	Chain       *ChainConfig
	P2P         *p2p.Config
	Shuttermint *kprconfig.ShuttermintConfig
	Metrics     *metricsserver.MetricsConfig

	MaxNumKeysPerMessage uint64
}

func (c *Config) Validate() error {
	//TODO: needs to be implemented
	return nil
}

func (c *Config) Name() string {
	return "shutterservice"
}

func (c *Config) SetDefaultValues() error {
	c.HTTPEnabled = false
	c.HTTPListenAddress = ":3000"
	c.MaxNumKeysPerMessage = 500 //TODO: need to confirm on this
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
	return c.Chain.Node.PrivateKey.EthereumAddress()
}

type ChainConfig struct {
	Node                 *configuration.EthnodeConfig `shconfig:",required"`
	Contracts            *ContractsConfig             `shconfig:",required"`
	SyncStartBlockNumber uint64                       `shconfig:",required"`
}

func NewChainConfig() *ChainConfig {
	c := &ChainConfig{
		Node:                 configuration.NewEthnodeConfig(),
		Contracts:            NewContractsConfig(),
		SyncStartBlockNumber: 0,
	}
	c.Init()
	return c
}

func (c *ChainConfig) Init() {
	c.Node.Init()
	c.Contracts.Init()
}

func (c *ChainConfig) Name() string {
	return "chain"
}

func (c *ChainConfig) Validate() error {
	return nil
}

func (c *ChainConfig) SetDefaultValues() error {
	c.SyncStartBlockNumber = 0
	return nil
}

func (c *ChainConfig) SetExampleValues() error {
	return nil
}

func (c *ChainConfig) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}

type ContractsConfig struct {
	KeyperSetManager     common.Address `shconfig:",required"`
	ShutterRegistry      common.Address `shconfig:",required"`
	KeyBroadcastContract common.Address `shconfig:",required"`
}

func NewContractsConfig() *ContractsConfig {
	return &ContractsConfig{
		KeyperSetManager:     common.Address{},
		ShutterRegistry:      common.Address{},
		KeyBroadcastContract: common.Address{},
	}
}

func (c *ContractsConfig) Init() {}

func (c *ContractsConfig) Name() string {
	return "shutterservicecontracts"
}

func (c *ContractsConfig) Validate() error {
	return nil
}

func (c *ContractsConfig) SetDefaultValues() error {
	return nil
}

func (c *ContractsConfig) SetExampleValues() error {
	return nil
}

func (c *ContractsConfig) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}
