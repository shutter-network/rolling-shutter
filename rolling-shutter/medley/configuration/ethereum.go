package configuration

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/keys"
)

var _ Config = &EthnodeConfig{}

func NewEthnodeConfig() *EthnodeConfig {
	c := &EthnodeConfig{}
	c.Init()
	return c
}

type EthnodeConfig struct {
	PrivateKey    *keys.ECDSAPrivate `shconfig:",required"`
	ContractsURL  string             `                     comment:"The JSON RPC endpoint where the contracts are accessible"`
	DeploymentDir string             `                     comment:"Contract source directory"`
	EthereumURL   string             `                     comment:"The layer 1 JSON RPC endpoint"`
}

func (c *EthnodeConfig) Init() {
	c.PrivateKey = &keys.ECDSAPrivate{}
}

func (c *EthnodeConfig) Name() string {
	return "ethnode"
}

func (c *EthnodeConfig) Validate() error {
	return nil
}

func (c *EthnodeConfig) SetDefaultValues() error {
	c.EthereumURL = "http://127.0.0.1:8545/"
	c.ContractsURL = "http://127.0.0.1:8555/"
	c.DeploymentDir = "./deployments/localhost/"
	return nil
}

func (c *EthnodeConfig) SetExampleValues() error {
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

func (c EthnodeConfig) TOMLWriteHeader(w io.Writer) (int, error) {
	return fmt.Fprintf(w, "# Ethereum address: %s\n", c.PrivateKey.EthereumAddress())
}
