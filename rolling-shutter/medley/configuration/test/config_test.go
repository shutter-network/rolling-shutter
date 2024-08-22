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

// nolint: lll, nolintlint
const expectedToml = `# first header
# second header


UInt64 = 123
Bool = true
String = 'barfoo'

[NestedConfig]
NestedDefaultInt = 42
NestedDefaultString = 'parentfoobar'
NestedEd25519Public = '0ac077275039117a76cc6f70a27e721c50ee3bcf8163efe273dc19b87cbebf87'
NestedEd25519Private = '500418849232a0c918b6333450b4afab775101eb1504fd94fe46e1e6fd53a2c2'
NestedECDSAPublic = '04dd5e32fe9fc5db272cca9f2bdeb6f7f9bae36258cd51ae8edd783f9cb7a60e16a4ce1834702c1e031da9f452356bea7f96a06ddcc8bd1b80c61536a90d6454ce'
NestedECDSAPrivate = '8a1e84264bd35f11ec867ebc8328cf0b29e427cc1d38b63919339e291b6863ac'
NestedLibp2pPrivate = 'CAESQOwP/OT9r6+HApehrPIN5hA/zeKfw2HWucEuRm3mVXfAEUE5d/MkEYCyakQwIbirosHMoSC7MH3p+vic4yrLyVc='
NestedLibp2pPublic = 'CAESIBFBOXfzJBGAsmpEMCG4q6LBzKEguzB96fr4nOMqy8lX'
`

func TestConfiguration(t *testing.T) {
	config := NewConfig()
	err := configuration.SetExampleValuesRecursive(config)
	assert.NilError(t, err)

	v := viper.New()
	afs := afero.NewMemMapFs()
	v.SetFs(afs)

	configFile := dirPath + config.Name() + "toml"
	err = afs.MkdirAll(dirPath, os.ModeDir)
	assert.NilError(t, err)

	err = command.WriteConfig(afs, config, configFile, false)
	assert.NilError(t, err)

	file, err := afero.ReadFile(afs, configFile)
	assert.NilError(t, err)

	assert.Equal(t, string(file), expectedToml)

	parsedConfig := NewConfig()

	v.SetFs(afs)
	v.SetConfigName(config.Name())
	v.SetConfigType("toml")
	v.SetConfigFile(configFile)
	err = v.ReadInConfig()
	assert.NilError(t, err)

	err = command.ParseViper(v, parsedConfig)
	assert.NilError(t, err)
	assert.DeepEqual(t, config, parsedConfig)
}

func TestSmokeGenerateConfig(t *testing.T) {
	config := NewConfig()
	SmokeGenerateConfig(t, config)
}
