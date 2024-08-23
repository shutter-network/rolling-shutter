package test

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/configuration/command"
)

const dirPath = "/test/config/"

func RoundtripParseConfig[T configuration.Config](
	t *testing.T,
	config T,
) T {
	t.Helper()

	v := viper.New()
	afs := afero.NewMemMapFs()
	v.SetFs(afs)

	configFile := dirPath + config.Name() + ".toml"
	err := afs.MkdirAll(dirPath, os.ModeDir)
	assert.NilError(t, err)

	err = command.WriteConfig(afs, config, configFile, false)
	assert.NilError(t, err)

	var parsedConfig T
	captureNewConfig := func(cfg T) error {
		parsedConfig = cfg
		return nil
	}

	cmd := command.Build(
		captureNewConfig,
		command.WithFileSystem(afs),
	).Command()
	cmd.SetArgs([]string{"--config", configFile})
	err = cmd.Execute()
	assert.NilError(t, err)
	return parsedConfig
}

// SmokeGenerateConfig is a basic smoketest to check that
// the SetExampleValues, WriteTOML, ParseTOML pipeline as well as
// the command flags and the command builder are sane and not
// panicking etc.
func SmokeGenerateConfig[T configuration.Config](
	t *testing.T,
	config T,
) {
	t.Helper()

	v := viper.New()
	afs := afero.NewMemMapFs()
	v.SetFs(afs)

	configFile := dirPath + config.Name() + ".toml"
	err := afs.MkdirAll(dirPath, os.ModeDir)
	assert.NilError(t, err)

	mainTest := func(cfg T) error {
		return nil
	}

	cmd := command.Build(
		mainTest,
		command.WithGenerateConfigSubcommand(),
		command.WithFileSystem(afs),
	).Command()

	cmd.SetArgs([]string{"generate-config", "--output", configFile})
	err = cmd.Execute()
	assert.NilError(t, err)

	cmd.SetArgs([]string{"--config", configFile})
	err = cmd.Execute()
	assert.NilError(t, err)
}
