package collator_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/collator/config"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/test"
)

func TestSmokeGenerateConfig(t *testing.T) {
	cfg := config.New()
	test.SmokeGenerateConfig(t, cfg)
}

func TestParsedConfig(t *testing.T) {
	cfg := config.New()

	err := configuration.SetExampleValuesRecursive(cfg)
	assert.NilError(t, err)
	parsedConfig := test.RoundtripParseConfig(t, cfg)
	assert.DeepEqual(t, cfg, parsedConfig)
}
