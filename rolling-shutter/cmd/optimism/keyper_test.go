package optimism_test

import (
	"testing"

	"gotest.tools/assert"

	config "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/test"
)

func TestSmokeGenerateConfig(t *testing.T) {
	c := config.NewConfig()
	test.SmokeGenerateConfig(t, c)
}

func TestParsedConfig(t *testing.T) {
	c := config.NewConfig()

	err := configuration.SetExampleValuesRecursive(c)
	assert.NilError(t, err)
	parsedConfig := test.RoundtripParseConfig(t, c)
	assert.DeepEqual(t, c, parsedConfig)
}
