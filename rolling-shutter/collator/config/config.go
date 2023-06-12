package config

import (
	"io"
	"time"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/epoch"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

var _ configuration.Config = &Config{}

func New() *Config {
	c := &Config{}
	c.Init()
	return c
}

func (c *Config) Init() {
	c.P2P = p2p.NewConfig()
	c.Ethereum = configuration.NewEthnodeConfig()
}

type Config struct {
	InstanceID  uint64 `shconfig:",required"`
	DatabaseURL string `shconfig:",required"`

	HTTPListenAddress string

	SequencerURL                 string
	EpochDuration                *epoch.Duration
	ExecutionBlockDelay          uint32
	BatchIndexAcceptenceInterval uint32

	P2P      *p2p.Config
	Ethereum *configuration.EthnodeConfig
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Name() string {
	return "collator"
}

func (c *Config) SetDefaultValues() error {
	c.EpochDuration = &epoch.Duration{
		Duration: time.Second * 5,
	}
	c.SequencerURL = "http://127.0.0.1:8555/"
	// default: the contracts are deployed on L2
	c.Ethereum.ContractsURL = c.SequencerURL
	c.BatchIndexAcceptenceInterval = 5
	c.ExecutionBlockDelay = 5
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
