package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"gotest.tools/assert"

	"github.com/shutter-network/shutter/shuttermint/keyper"
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
	_ = tomlRoundtrip(t, cfg)
}
