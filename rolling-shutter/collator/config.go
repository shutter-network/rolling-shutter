package collator

import (
	"crypto/ecdsa"
	"io"
	"text/template"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/medley"
)

type Config struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr

	EthereumURL   string
	DeploymentDir string
	DatabaseURL   string

	EthereumKey *ecdsa.PrivateKey
	P2PKey      libp2pcrypto.PrivKey

	InstanceID uint64
}

var configTemplate = `# Shutter collator config
# Ethereum address: {{ .EthereumAddress }}
# Peer identity: /p2p/{{ .P2PKey | P2PKeyPublic}}

EthereumURL     = "{{ .EthereumURL }}"
DeploymentDir   = "{{ .DeploymentDir }}"

# DatabaseURL looks like postgres://username:password@localhost:5432/database_name
# If it's empty, we use the standard PG* environment variables
DatabaseURL     = "{{ .DatabaseURL }}"

# p2p configuration
ListenAddress   = "{{ .ListenAddress }}"
PeerMultiaddrs  = [{{ .PeerMultiaddrs | QuoteList}}]

# Secret Keys
EthereumKey     = "{{ .EthereumKey | FromECDSA | printf "%x" }}"
P2PKey          = "{{ .P2PKey | P2PKey}}"

# ID shared by all shutter participants for common instance
InstanceID = {{ .InstanceID }}
`

var tmpl *template.Template = medley.MustBuildTemplate("collator", configTemplate)

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
				medley.StringToEcdsaPrivateKey,
			),
		),
	)
}

func (config *Config) EthereumAddress() common.Address {
	return ethcrypto.PubkeyToAddress(config.EthereumKey.PublicKey)
}
