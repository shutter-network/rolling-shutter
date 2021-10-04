package decryptor

import (
	"io"
	"text/template"

	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/medley"
)

type Config struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr

	DatabaseURL string

	P2PKey     crypto.PrivKey
	SigningKey *shbls.SecretKey

	RequiredSignatures uint

	InstanceID uint64
}

var configTemplate = `# Shutter decryptor config for /p2p/{{ .P2PKey | P2PKeyPublic}}

# DatabaseURL looks like postgres://username:password@localhost:5432/database_name
# It it's empty, we use the standard PG* environment variables
DatabaseURL     = "{{ .DatabaseURL }}"

# p2p configuration
ListenAddress   = "{{ .ListenAddress }}"
PeerMultiaddrs  = [{{ .PeerMultiaddrs | QuoteList}}]

# Secret Keys
P2PKey          = "{{ .P2PKey | P2PKey}}"
SigningKey      = "{{ .SigningKey | BLSSecretKey}}"

# Number of individual signatures required to form an accepted aggregated signature
requiredSignatures = {{.RequiredSignatures}}

# ID shared by all shutter participants for common instance
InstanceID = {{ .InstanceID }}
`

var tmpl *template.Template = medley.MustBuildTemplate("decryptor", configTemplate)

// WriteTOML writes a toml configuration file with the given config.
func (config *Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}

// Unmarshal unmarshals a DecryptorConfig from the given Viper object.
func (config *Config) Unmarshal(v *viper.Viper) error {
	return v.Unmarshal(
		config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				medley.MultiaddrHook,
				medley.P2PKeyHook,
				medley.BLSSecretKeyHook,
			),
		),
	)
}
