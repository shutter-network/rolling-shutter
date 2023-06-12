package keyper

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/dkgphase"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/keys"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

var (
	_ configuration.Config = &ShuttermintConfig{}
	_ configuration.Config = &Config{}
)

func NewConfig() *Config {
	c := &Config{}
	c.Init()
	return c
}

func (c *Config) Init() {
	c.P2P = p2p.NewConfig()
	c.Ethereum = configuration.NewEthnodeConfig()
	c.Shuttermint = NewShuttermintConfig()
}

type Config struct {
	InstanceID  uint64 `shconfig:",required"`
	DatabaseURL string `shconfig:",required" comment:"If it's empty, we use the standard PG_ environment variables"`

	HTTPEnabled       bool
	HTTPListenAddress string

	P2P         *p2p.Config
	Ethereum    *configuration.EthnodeConfig
	Shuttermint *ShuttermintConfig
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
	return ecies.ImportECDSA(c.Ethereum.PrivateKey.Key)
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

func NewShuttermintConfig() *ShuttermintConfig {
	c := &ShuttermintConfig{}
	c.Init()
	return c
}

type ShuttermintConfig struct {
	ShuttermintURL     string
	ValidatorPublicKey *keys.Ed25519Public `shconfig:",required"`
	DKGPhaseLength     int64               // in shuttermint blocks
	DKGStartBlockDelta uint64
}

func (c *ShuttermintConfig) Init() {
	c.ValidatorPublicKey = &keys.Ed25519Public{}
}

func (c *ShuttermintConfig) Name() string {
	return "shuttermint"
}

func (c *ShuttermintConfig) Validate() error {
	if c.DKGPhaseLength < 0 {
		return errors.New("DKGPhaseLength can't be negative")
	}
	return nil
}

func (c *ShuttermintConfig) SetDefaultValues() error {
	c.ShuttermintURL = "http://localhost:26657"
	c.DKGPhaseLength = 30
	c.DKGStartBlockDelta = 200
	return nil
}

func (c *ShuttermintConfig) SetExampleValues() error {
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}

	valPriv, err := keys.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return err
	}
	c.ValidatorPublicKey = valPriv.Public().(*keys.Ed25519Public)
	return nil
}

func (c ShuttermintConfig) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}
