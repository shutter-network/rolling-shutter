package bootstrap

import (
	"crypto/rand"
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/keys"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

var _ configuration.Config = &Config{}

func NewConfig() *Config {
	c := &Config{}
	c.Init()
	return c
}

func (c *Config) Init() {
	c.SigningKey = &keys.ECDSAPrivate{}
	c.ActivationBlockNumber = number.NewBlockNumber()
}

type Config struct {
	InstanceID uint64 `shconfig:",required"`

	JSONRPCURL            string `                     comment:"The op-geth JSON RPC endpoint"`
	ActivationBlockNumber *number.BlockNumber
	KeyperSetFilePath     string

	ShuttermintURL string
	SigningKey     *keys.ECDSAPrivate `shconfig:",required"`
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Name() string {
	return "op-bootstrap"
}

func (c *Config) SetDefaultValues() error {
	c.JSONRPCURL = "http://localhost:8545"
	c.ShuttermintURL = "http://localhost:26657"
	c.KeyperSetFilePath = "keyperset.json"
	c.ActivationBlockNumber = number.LatestBlock
	return nil
}

func (c *Config) SetExampleValues() error {
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}
	c.SigningKey, err = keys.GenerateECDSAKey(rand.Reader)
	if err != nil {
		return err
	}
	c.InstanceID = 42
	return nil
}

func (c Config) TOMLWriteHeader(_ io.Writer) (int, error) {
	return 0, nil
}