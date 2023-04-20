package snapshot

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io"
	"text/template"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
)

type Config struct {
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr

	EthereumURL    string
	DatabaseURL    string
	SnapshotHubURL string

	JSONRPCHost string
	JSONRPCPort uint16

	MetricsEnabled bool
	MetricsHost    string
	MetricsPort    uint16

	EthereumKey *ecdsa.PrivateKey
	P2PKey      p2pcrypto.PrivKey
}

const configTemplate = `# Shutter snapshot config
# Ethereum address: {{ .EthereumAddress }}
# Peer identity: /p2p/{{ .P2PKey | P2PKeyPublic}}

EthereumURL     = "{{ .EthereumURL }}"

# DatabaseURL looks like postgres://username:password@localhost:5432/database_name
# If it's empty, we use the standard PG* environment variables
DatabaseURL     = "{{ .DatabaseURL }}"

# Snapshot integration
SnapshotHubURL       = "{{ .SnapshotHubURL }}"

# JSONRPC configuration
JSONRPCHost     = "{{ .JSONRPCHost }}"
JSONRPCPort     = {{ .JSONRPCPort }}

# Metrics configuration
MetricsEnabled  = {{ .MetricsEnabled }}
MetricsHost     = "{{ .MetricsHost }}"
MetricsPort     = {{ .MetricsPort }}

# p2p configuration
ListenAddress   = "{{ .ListenAddress }}"
PeerMultiaddrs  = [{{ .PeerMultiaddrs | QuoteList}}]

# Secret Keys
EthereumKey     = "{{ .EthereumKey | FromECDSA | printf "%x" }}"
P2PKey          = "{{ .P2PKey | P2PKey}}"
`

var tmpl *template.Template = medley.MustBuildTemplate("snapshot", configTemplate)

func (config *Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}

// Unmarshal unmarshals a SnapshotConfig from the given Viper object.
func (config *Config) Unmarshal(v *viper.Viper) error {
	err := v.Unmarshal(
		config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				medley.MultiaddrHook,
				medley.P2PKeyHook,
				medley.StringToEcdsaPrivateKey,
				mapstructure.StringToTimeDurationHookFunc(),
			),
		),
	)
	if err != nil {
		return err
	}
	if config.EthereumKey == nil {
		return errors.Errorf("EthereumKey is missing")
	}
	return nil
}

func (config *Config) EthereumAddress() common.Address {
	return ethcrypto.PubkeyToAddress(config.EthereumKey.PublicKey)
}
