package p2pnode_test

import (
	"testing"

	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/test"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2pnode"
)

func TestSmokeGenerateConfig(t *testing.T) {
	config := p2pnode.NewConfig()
	test.SmokeGenerateConfig(t, config)
}

func TestParsedConfig(t *testing.T) {
	config := p2pnode.NewConfig()

	err := configuration.SetExampleValuesRecursive(config)
	assert.NilError(t, err)
	parsedConfig := test.RoundtripParseConfig(t, config)
	assert.DeepEqual(t, config, parsedConfig)
}
