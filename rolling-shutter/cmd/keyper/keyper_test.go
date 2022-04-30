package keyper

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shlib/shtest"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/comparer"
)

func tomlRoundtrip(t *testing.T, cfg *keyper.Config) *keyper.Config {
	t.Helper()
	var buf bytes.Buffer
	err := cfg.WriteTOML(&buf)
	assert.NilError(t, err)

	fmt.Println(buf.String())

	v := viper.New()
	v.SetConfigType("toml")

	err = v.ReadConfig(&buf)
	assert.NilError(t, err)

	cfg2 := &keyper.Config{}
	err = cfg2.Unmarshal(v)
	assert.NilError(t, err)
	return cfg2
}

func TestGeneratedConfigValid(t *testing.T) {
	cfg, err := exampleConfig()
	assert.NilError(t, err)
	cfg2 := tomlRoundtrip(t, cfg)
	assert.DeepEqual(t, cfg, cfg2,
		shtest.BigIntComparer,
		comparer.P2PPrivKeyComparer,
		comparer.EciesPrivateKeyComparer,
	)
}
