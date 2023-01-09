package mocksequencer

import (
	"io"
	"text/template"

	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
)

var configTemplate = `# Shutter mock sequencer config
# JSON-RPC endpoint exposed to clients
HTTPListenAddress = "{{ .HTTPListenAddress }}"

# Layer1 JSON-RPC endpoint
EthereumURL = "{{ .EthereumURL }}"

# Chain config
ChainID = {{ .ChainID }}
MaxBlockDeviation = {{ .MaxBlockDeviation }}
EthereumPollInterval = {{ .EthereumPollInterval }}

# Debug
Admin = {{ .Admin }}
Debug = {{ .Debug }}

`

var tmpl *template.Template = medley.MustBuildTemplate("mocksequencer", configTemplate)

type Config struct {
	HTTPListenAddress string
	EthereumURL       string

	ChainID              uint64
	MaxBlockDeviation    uint64
	EthereumPollInterval uint64

	Admin bool
	Debug bool
}

// WriteTOML writes a toml configuration file with the given config.
func (config *Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}

// Unmarshal unmarshals a mocksequencer Config from the the given Viper object.
func (config *Config) Unmarshal(v *viper.Viper) error {
	return v.Unmarshal(config)
}
