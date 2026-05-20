package kprconfig

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/metricsserver"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type Config struct {
	InstanceID  uint64
	DatabaseURL string

	HTTPEnabled       bool
	HTTPReadOnly      bool
	HTTPListenAddress string

	P2P      *p2p.Config
	Ethereum *configuration.EthnodeConfig
	Metrics  *metricsserver.MetricsConfig

	MaxNumKeysPerMessage uint64
}

func (c *Config) GetAddress() common.Address {
	return c.Ethereum.PrivateKey.EthereumAddress()
}

func (c *Config) GetInstanceID() uint64 {
	return c.InstanceID
}

func (c *Config) GetHTTPListenAddress() string {
	return c.HTTPListenAddress
}

func (c *Config) GetEnableWriteOperations() bool {
	return c.HTTPEnabled && !c.HTTPReadOnly
}

func (c *Config) GetMaxNumKeysPerMessage() uint64 {
	return c.MaxNumKeysPerMessage
}
