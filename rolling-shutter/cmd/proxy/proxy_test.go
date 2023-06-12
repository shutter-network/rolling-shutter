package proxy_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/test"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/proxy"
)

func TestSmokeGenerateConfig(t *testing.T) {
	config := proxy.NewConfig()
	test.SmokeGenerateConfig(t, config)
}

func TestParsedConfig(t *testing.T) {
	config := proxy.NewConfig()

	err := configuration.SetExampleValuesRecursive(config)
	assert.NilError(t, err)
	parsedConfig := test.RoundtripParseConfig(t, config)
	assert.DeepEqual(t, config, parsedConfig)
}
