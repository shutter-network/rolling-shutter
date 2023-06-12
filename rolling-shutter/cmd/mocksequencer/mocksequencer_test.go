package mocksequencer_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/test"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/mocksequencer"
)

func TestSmokeGenerateConfig(t *testing.T) {
	config := mocksequencer.NewConfig()
	test.SmokeGenerateConfig(t, config)
}

func TestParsedConfig(t *testing.T) {
	config := mocksequencer.NewConfig()

	err := configuration.SetExampleValuesRecursive(config)
	assert.NilError(t, err)
	parsedConfig := test.RoundtripParseConfig(t, config)
	assert.DeepEqual(t, config, parsedConfig)
}
