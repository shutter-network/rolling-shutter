package mocknode

import (
	"io"
	"text/template"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/medley"
)

type Config struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	P2PKey         crypto.PrivKey

	InstanceID             uint64
	Rate                   float64
	SendDecryptionTriggers bool
	SendCipherBatches      bool
	SendDecryptionKeys     bool
}

var configTemplate = `# Shutter mock node config for /p2p/{{ .P2PKey | P2PKeyPublic}}

# p2p configuration
ListenAddress   = "{{ .ListenAddress }}"
PeerMultiaddrs  = [{{ .PeerMultiaddrs | QuoteList}}]

# Secret Keys
P2PKey          = "{{ .P2PKey | P2PKey}}"

# Mock messages
InstanceID              = {{ .InstanceID }}
Rate                    = {{ .Rate }}
SendDecryptionTriggers  = {{ .SendDecryptionTriggers }}
SendCipherBatches       = {{ .SendCipherBatches }}
SendDecryptionKeys      = {{ .SendDecryptionKeys }}
`

var tmpl *template.Template = medley.MustBuildTemplate("keyper", configTemplate)

// Unmarshal unmarshals a keyper Config from the the given Viper object.
func (config *Config) Unmarshal(v *viper.Viper) error {
	return v.Unmarshal(
		config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				medley.MultiaddrHook,
				medley.P2PKeyHook,
			),
		),
	)
}

// WriteTOML writes a toml configuration file with the given config.
func (config *Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}
