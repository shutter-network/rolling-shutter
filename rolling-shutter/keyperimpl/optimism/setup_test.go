package optimism

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/testsetup"
)

func init() {
	var err error
	testConfig.collatorKey, err = ethcrypto.GenerateKey()
	if err != nil {
		panic(errors.Wrap(err, "ethcrypto.GenerateKey failed"))
	}
}

type TestConfig struct {
	collatorKey *ecdsa.PrivateKey
}

var testConfig = &TestConfig{}

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
	return testConfig.collatorKey
}

var _ testsetup.TestConfig = &TestConfig{}
