package config

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/keys"
)

var _ configuration.Config = &OptimismConfig{}

func NewEthnodeConfig() *OptimismConfig {
	c := &OptimismConfig{}
	c.Init()
	return c
}

type OptimismConfig struct {
	PrivateKey *keys.ECDSAPrivate `shconfig:",required"`
	JSONRPCURL string             `                     comment:"The op-geth JSON RPC endpoint"`
}

func (c *OptimismConfig) Init() {
	c.PrivateKey = &keys.ECDSAPrivate{}
}

func (c *OptimismConfig) Name() string {
	return "ethnode"
}

func (c *OptimismConfig) Validate() error {
	return nil
}

func (c *OptimismConfig) SetDefaultValues() error {
	c.JSONRPCURL = "http://127.0.0.1:8545/"
	return nil
}

func (c *OptimismConfig) SetExampleValues() error {
	err := c.SetDefaultValues()
	if err != nil {
		return err
	}

	c.PrivateKey, err = keys.GenerateECDSAKey(rand.Reader)
	if err != nil {
		return err
	}
	return nil
}

func (c OptimismConfig) TOMLWriteHeader(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "# Ethereum address: %s\n", c.PrivateKey.EthereumAddress())
}
