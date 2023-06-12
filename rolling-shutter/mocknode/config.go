package mocknode

import (
	"io"

	"github.com/shutter-network/shutter/shlib/shcrypto"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
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
}

type Config struct {
	InstanceID uint64 `shconfig:",required"`
	EonKeySeed int64  `shconfig:",required" comment:"a seed value used to generate the eon key"`

	Rate                   float64 `comment:"overall rate (in seconds) influencing tx send frequency"`
	SendDecryptionTriggers bool
	SendDecryptionKeys     bool
	SendTransactions       bool

	P2P      *p2p.Config
	Ethereum *configuration.EthnodeConfig
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Name() string {
	return "mocknode"
}

func (c *Config) SetDefaultValues() error {
	c.Rate = 1.0
	c.SendDecryptionTriggers = true
	c.SendDecryptionKeys = true
	c.SendTransactions = true
	return nil
}

func (c *Config) SetExampleValues() error {
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}
	c.InstanceID = 42
	c.EonKeySeed = 1337
	return nil
}

func (c Config) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}

// EonPublicKey returns the eon public key defined by the seed value in the config.
func (c *Config) EonPublicKey() *shcrypto.EonPublicKey {
	_, eonPublicKey, err := computeEonKeys(c.EonKeySeed)
	if err != nil {
		panic(err)
	}
	return eonPublicKey
}
