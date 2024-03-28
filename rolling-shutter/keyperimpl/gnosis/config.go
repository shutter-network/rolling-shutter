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

	// TODO: put these in a child config
	GnosisContracts      *GnosisContracts `shconfig:",required"`
	EncryptedGasLimit    uint64           `shconfig:",required"`
	MinGasPerTransaction uint64           `shconfig:",required"`
	SecondsPerSlot       uint64           `shconfig:",required"`
	GenesisSlotTimestamp uint64           `shconfig:",required"`
}

type GnosisContracts struct {
	KeyperSetManager     common.Address `shconfig:",required"`
	KeyBroadcastContract common.Address `shconfig:",required"`
	Sequencer            common.Address `shconfig:",required"`
}

func (c *Config) Validate() error {
	if c.SecondsPerSlot > maxSecondsPerSlot {
		return errors.Errorf("seconds per slot is too big (%d > %d)", c.SecondsPerSlot, maxSecondsPerSlot)
	}
	maxGenesisSlotTime := uint64(math.MaxInt64 - maxChainAge)
	if c.GenesisSlotTimestamp > maxGenesisSlotTime {
		return errors.Errorf("genesis slot timestamp is too big (%d > %d)", c.GenesisSlotTimestamp, maxGenesisSlotTime)
	}
	return nil
}

func (c *Config) Name() string {
	return "gnosiskeyper"
}

func (c *Config) SetDefaultValues() error {
	c.HTTPEnabled = false
	c.HTTPListenAddress = ":3000"
	c.GnosisContracts = &GnosisContracts{
		KeyperSetManager:     common.Address{},
		KeyBroadcastContract: common.Address{},
		Sequencer:            common.Address{},
	}
	c.EncryptedGasLimit = 1_000_000
	c.MinGasPerTransaction = 21_000
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
