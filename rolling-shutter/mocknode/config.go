package mocknode

import (
	"io"
	"text/template"

	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shlib/shcrypto"
	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
	"github.com/shutter-network/shutter/shuttermint/medley"
)

type Config struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
	P2PKey         p2pcrypto.PrivKey

	InstanceID             uint64
	Rate                   float64
	SendDecryptionTriggers bool
	SendCipherBatches      bool
	SendDecryptionKeys     bool
	SendTransactions       bool

	DecryptorPublicKeys []*shbls.PublicKey // public keys of decryptors in order of their index
	EonKeySeed          int64              // a seed value used to generate the eon key

}

var configTemplate = `# Shutter mock node config
# Peer Identity: /p2p/{{ .P2PKey | P2PKeyPublic}}
# Eon Public Key: {{ .EonPublicKey | EonPublicKey }}

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
SendTransactions        = {{ .SendTransactions }}

DecryptorPublicKeys = {{ .DecryptorPublicKeys | BLSPublicKeys }}
EonKeySeed         = {{ .EonKeySeed }}
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
				medley.BLSPublicKeyHook,
			),
		),
	)
}

// WriteTOML writes a toml configuration file with the given config.
func (config *Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}

// EonPublicKey returns the eon public key defined by the seed value in the config.
func (config *Config) EonPublicKey() *shcrypto.EonPublicKey {
	_, eonPublicKey, err := computeEonKeys(config.EonKeySeed)
	if err != nil {
		panic(err)
	}
	return eonPublicKey
}
