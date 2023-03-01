package p2pnode

import (
	"crypto/rand"
	"io"
	"text/template"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type Config struct {
	PrivateKey               crypto.PrivKey
	ListenAddresses          []multiaddr.Multiaddr
	CustomBootstrapAddresses []peer.AddrInfo
	Environment              p2p.Environment
}

const configTemplate = `# Shutter  p2p node config
# Peer role: bootstrap
# Peer identity: /p2p/{{ .PrivateKey | P2PKeyPublic}}

# p2p configuration
ListenAddresses   = [{{ .ListenAddresses | QuoteList}}]
CustomBootstrapAddresses  = [{{ .CustomBootstrapAddresses | ToMultiAddrList | QuoteList}}]

# Secret Keys
PrivateKey          = "{{ .PrivateKey | P2PKey}}"

`

var tmpl *template.Template = medley.MustBuildTemplate("p2pnode", configTemplate)

// GenerateNewKeys generates new keys and stores them inside the Config object.
func (config *Config) GenerateNewKeys() error {
	p2pkey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return err
	}
	config.PrivateKey = p2pkey
	return nil
}

// Unmarshal unmarshals a keyper Config from the the given Viper object.
func (config *Config) Unmarshal(v *viper.Viper) error {
	err := v.Unmarshal(
		config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				medley.StringToEnvironment,
				medley.StringToEd25519PrivateKey,
				medley.StringToEd25519PublicKey,
				medley.P2PKeyHook,
				mapstructure.StringToSliceHookFunc(","),
				medley.MultiaddrHook,
				medley.AddrInfoHook,
			),
		),
	)
	if err != nil {
		return err
	}
	return nil
}

// WriteTOML writes a toml configuration file with the given config.
func (config *Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}
