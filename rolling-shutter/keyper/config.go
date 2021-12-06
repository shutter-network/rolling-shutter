package keyper

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"text/template"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/medley"
)

// Config contains validated configuration parameters for the keyper client.
type Config struct {
	ShuttermintURL string
	EthereumURL    string
	DatabaseURL    string
	DeploymentDir  string

	SigningKey     *ecdsa.PrivateKey
	ValidatorKey   ed25519.PrivateKey `mapstructure:"ValidatorSeed"`
	EncryptionKey  *ecies.PrivateKey
	P2PKey         p2pcrypto.PrivKey
	DKGPhaseLength uint64 // in shuttermint blocks
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr

	InstanceID uint64
}

const configTemplate = `# Shutter keyper config
# Ethereum address: {{ .Address }}
# Peer identity: /p2p/{{ .P2PKey | P2PKeyPublic}}

ShuttermintURL		= "{{ .ShuttermintURL }}"
EthereumURL         = "{{ .EthereumURL }}"
DeploymentDir       = "{{ .DeploymentDir }}"

# DatabaseURL looks like postgres://username:password@localhost:5432/database_name
# If it's empty, we use the standard PG* environment variables
DatabaseURL		= "{{ .DatabaseURL }}"
DKGPhaseLength		= {{ .DKGPhaseLength }}

# p2p configuration
ListenAddress	= "{{ .ListenAddress }}"
PeerMultiaddrs	= [{{ .PeerMultiaddrs | QuoteList}}]

# Secret Keys
EncryptionKey	= "{{ .EncryptionKey.ExportECDSA | FromECDSA | printf "%x" }}"
SigningKey	= "{{ .SigningKey | FromECDSA | printf "%x" }}"
ValidatorSeed	= "{{ .ValidatorKey.Seed | printf "%x" }}"
P2PKey          = "{{ .P2PKey | P2PKey}}"

InstanceID = {{ .InstanceID }}
`

var tmpl *template.Template = medley.MustBuildTemplate("keyper", configTemplate)

func randomSigningKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateKey()
}

func randomEncryptionKey() (*ecies.PrivateKey, error) {
	encryptionKeyECDSA, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return ecies.ImportECDSA(encryptionKeyECDSA), nil
}

func randomValidatorKey() (ed25519.PrivateKey, error) {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		return nil, err
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

// GenerateNewKeys generates new keys and stores them inside the Config object.
func (config *Config) GenerateNewKeys() error {
	signingKey, err := randomSigningKey()
	if err != nil {
		return err
	}
	encryptionKey, err := randomEncryptionKey()
	if err != nil {
		return err
	}
	validatorKey, err := randomValidatorKey()
	if err != nil {
		return err
	}

	p2pkey, _, err := p2pcrypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return err
	}

	config.SigningKey = signingKey
	config.ValidatorKey = validatorKey
	config.EncryptionKey = encryptionKey
	config.P2PKey = p2pkey
	return nil
}

// Unmarshal unmarshals a keyper Config from the the given Viper object.
func (config *Config) Unmarshal(v *viper.Viper) error {
	err := v.Unmarshal(
		config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				medley.StringToEd25519PrivateKey,
				medley.StringToEcdsaPrivateKey,
				medley.StringToEciesPrivateKey,
				medley.StringToAddress,
				medley.P2PKeyHook,
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
				medley.MultiaddrHook,
			),
		),
	)
	if err != nil {
		return err
	}
	if config.SigningKey == nil {
		return errors.Errorf("SigningKey is missing")
	}
	if config.EncryptionKey == nil {
		return errors.Errorf("EncryptionKey is missing")
	}
	return nil
}

// Address returns the keyper's Ethereum address.
func (config *Config) Address() common.Address {
	return crypto.PubkeyToAddress(config.SigningKey.PublicKey)
}

// WriteTOML writes a toml configuration file with the given config.
func (config *Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}
