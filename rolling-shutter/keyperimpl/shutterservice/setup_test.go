package shutterservice

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

// kprAddressTestConfig is a TestConfig that substitutes a specific address
// while delegating all other fields to the global config. Used to register
// a dynamically-generated keyper address in InitializeEon.
type kprAddressTestConfig struct {
	addr common.Address
}

func (c *kprAddressTestConfig) GetAddress() common.Address        { return c.addr }
func (c *kprAddressTestConfig) GetInstanceID() uint64             { return config.GetInstanceID() }
func (c *kprAddressTestConfig) GetEon() uint64                    { return config.GetEon() }
func (c *kprAddressTestConfig) GetCollatorKey() *ecdsa.PrivateKey { return nil }
func (c *kprAddressTestConfig) GetMaxNumKeysPerMessage() uint64 {
	return config.GetMaxNumKeysPerMessage()
}

var _ testsetup.TestConfig = &kprAddressTestConfig{}
