package mocknode_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/test"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocknode"
)

func TestSmokeGenerateConfig(t *testing.T) {
	config := mocknode.NewConfig()
	test.SmokeGenerateConfig(t, config)
}

func TestParsedConfig(t *testing.T) {
	config := mocknode.NewConfig()

	err := configuration.SetExampleValuesRecursive(config)
	assert.NilError(t, err)
	parsedConfig := test.RoundtripParseConfig(t, config)
	assert.DeepEqual(t, config, parsedConfig)
}
