package rollup

import (
	"crypto/ed25519"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/dkgphase"
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
	c.Ethereum = configuration.NewEthnodeConfig()
	c.Shuttermint = kprconfig.NewShuttermintConfig()
	c.Metrics = metricsserver.NewConfig()
}

type Config struct {
	InstanceID  uint64 `shconfig:",required"`
	DatabaseURL string `shconfig:",required" comment:"If it's empty, we use the standard PG_ environment variables"`

	HTTPEnabled       bool
	HTTPListenAddress string

	P2P         *p2p.Config
	Ethereum    *configuration.EthnodeConfig
	Shuttermint *kprconfig.ShuttermintConfig
	Metrics     *metricsserver.MetricsConfig
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) GetAddress() common.Address {
	return c.Ethereum.PrivateKey.EthereumAddress()
}

func (c *Config) GetDKGPhaseLength() *dkgphase.PhaseLength {
	return dkgphase.NewConstantPhaseLength(c.Shuttermint.DKGPhaseLength)
}

func (c *Config) GetValidatorPublicKey() ed25519.PublicKey {
	return c.Shuttermint.ValidatorPublicKey.Key
}

func (c *Config) GetEncryptionKey() *ecies.PrivateKey {
	// OPTIM this could be cached, but it is only used for
	// 		eon DKG (rarely) and does not do any computation
	return ecies.ImportECDSA(c.Shuttermint.EncryptionKey.Key)
}

func (c *Config) GetInstanceID() uint64 {
	return c.InstanceID
}

func (c *Config) Name() string {
	return "keyper"
}

func (c *Config) GetHTTPListenAddress() string {
	return c.HTTPListenAddress
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
