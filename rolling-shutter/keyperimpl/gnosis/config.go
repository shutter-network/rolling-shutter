package gnosis

import (
	"io"
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/kprconfig"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

const (
	maxSecondsPerSlot = 60 * 60
	maxChainAge       = 100 * 365 * 24 * 60 * 60
)

var (
	_ configuration.Config = &Config{}
	_ configuration.Config = &GnosisConfig{}
	_ configuration.Config = &GnosisContractsConfig{}
)

func NewConfig() *Config {
	c := &Config{}
	c.Init()
	return c
}

func (c *Config) Init() {
	c.P2P = p2p.NewConfig()
	c.Gnosis = NewGnosisConfig()
	c.Shuttermint = kprconfig.NewShuttermintConfig()
	c.Metrics = metricsserver.NewConfig()
}

type Config struct {
	InstanceID   uint64 `shconfig:",required"`
	DatabaseURL  string `shconfig:",required" comment:"If it's empty, we use the standard PG_ environment variables"`
	BeaconAPIURL string `shconfig:",required"`

	HTTPEnabled       bool
	HTTPListenAddress string

	Gnosis      *GnosisConfig
	P2P         *p2p.Config
	Shuttermint *kprconfig.ShuttermintConfig
	Metrics     *metricsserver.MetricsConfig
}

func (c *Config) Validate() error {
	if c.Gnosis.SecondsPerSlot > maxSecondsPerSlot {
		return errors.Errorf("seconds per slot is too big (%d > %d)", c.Gnosis.SecondsPerSlot, maxSecondsPerSlot)
	}
	maxGenesisSlotTime := uint64(math.MaxInt64 - maxChainAge)
	if c.Gnosis.GenesisSlotTimestamp > maxGenesisSlotTime {
		return errors.Errorf("genesis slot timestamp is too big (%d > %d)", c.Gnosis.GenesisSlotTimestamp, maxGenesisSlotTime)
	}
	return nil
}

func (c *Config) Name() string {
	return "gnosiskeyper"
}

func (c *Config) SetDefaultValues() error {
	c.HTTPEnabled = false
	c.HTTPListenAddress = ":3000"
	c.Gnosis.EncryptedGasLimit = 1_000_000
	c.Gnosis.MinGasPerTransaction = 21_000
	return nil
}

func (c *Config) SetExampleValues() error {
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}
	c.InstanceID = 42
	c.DatabaseURL = "postgres://pguser:pgpassword@localhost:5432/shutter"
	c.BeaconAPIURL = "http://localhost:5052"

	return nil
}

func (c Config) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}

func (c *Config) GetAddress() common.Address {
	return c.Gnosis.Node.PrivateKey.EthereumAddress()
}

type GnosisConfig struct {
	Node                 *configuration.EthnodeConfig `shconfig:",required"`
	Contracts            *GnosisContractsConfig       `shconfig:",required"`
	EncryptedGasLimit    uint64                       `shconfig:",required"`
	MinGasPerTransaction uint64                       `shconfig:",required"`
	SecondsPerSlot       uint64                       `shconfig:",required"`
	GenesisSlotTimestamp uint64                       `shconfig:",required"`
}

func NewGnosisConfig() *GnosisConfig {
	c := &GnosisConfig{
		Node:                 configuration.NewEthnodeConfig(),
		Contracts:            NewGnosisContractsConfig(),
		EncryptedGasLimit:    0,
		MinGasPerTransaction: 0,
		SecondsPerSlot:       0,
		GenesisSlotTimestamp: 0,
	}
	c.Init()
	return c
}

func (c *GnosisConfig) Init() {
	c.Node.Init()
	c.Contracts.Init()
}

func (c *GnosisConfig) Name() string {
	return "gnosis"
}

func (c *GnosisConfig) Validate() error {
	if c.SecondsPerSlot == 0 {
		return errors.Errorf("seconds per slot must not be zero")
	}
	return nil
}

func (c *GnosisConfig) SetDefaultValues() error {
	return nil
}

func (c *GnosisConfig) SetExampleValues() error {
	c.EncryptedGasLimit = 1_000_000
	c.MinGasPerTransaction = 21_000
	c.SecondsPerSlot = 5
	c.GenesisSlotTimestamp = 1665410700
	return nil
}

func (c *GnosisConfig) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}

type GnosisContractsConfig struct {
	KeyperSetManager     common.Address `shconfig:",required"`
	KeyBroadcastContract common.Address `shconfig:",required"`
	EonKeyPublish        common.Address `shconfig:",required"`
	Sequencer            common.Address `shconfig:",required"`
	ValidatorRegistry    common.Address `shconfig:",required"`
}

func NewGnosisContractsConfig() *GnosisContractsConfig {
	return &GnosisContractsConfig{
		KeyperSetManager:     common.Address{},
		KeyBroadcastContract: common.Address{},
		EonKeyPublish:        common.Address{},
		Sequencer:            common.Address{},
		ValidatorRegistry:    common.Address{},
	}
}

func (c *GnosisContractsConfig) Init() {}

func (c *GnosisContractsConfig) Name() string {
	return "gnosiscontracts"
}

func (c *GnosisContractsConfig) Validate() error {
	return nil
}

func (c *GnosisContractsConfig) SetDefaultValues() error {
	return nil
}

func (c *GnosisContractsConfig) SetExampleValues() error {
	return nil
}

func (c *GnosisContractsConfig) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}
