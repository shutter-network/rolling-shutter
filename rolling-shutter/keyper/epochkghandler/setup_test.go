package epochkghandler

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

type TestConfig struct{}

var config = &TestConfig{}

func (TestConfig) GetAddress() common.Address {
	return common.HexToAddress("0x2222222222222222222222222222222222222222")
}

func (TestConfig) GetInstanceID() uint64 {
	return 55
}

func (TestConfig) GetEon() uint64 {
	return 22
}

func (c *TestConfig) GetCollatorKey() *ecdsa.PrivateKey {
	return nil
}

func (c *TestConfig) GetMaxNumKeysPerMessage() uint64 {
	return 1024
}

var _ testsetup.TestConfig = &TestConfig{}
