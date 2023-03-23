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
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/dkgphase"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

// Config contains validated configuration parameters for the keyper client.
type Config struct {
	ShuttermintURL string
	EthereumURL    string
	ContractsURL   string
	DatabaseURL    string
	DeploymentDir  string

	ValidatorPublicKey       ed25519.PublicKey
	SigningKey               *ecdsa.PrivateKey
	EncryptionKey            *ecies.PrivateKey
	P2PKey                   p2pcrypto.PrivKey
	DKGPhaseLength           uint64 // in shuttermint blocks
	DKGStartBlockDelta       int64
	ListenAddresses          []multiaddr.Multiaddr
	CustomBootstrapAddresses []peer.AddrInfo
	Environment              p2p.Environment

	HTTPEnabled       bool
	HTTPListenAddress string

	InstanceID uint64
}

const configTemplate = `# Shutter keyper config
# Ethereum address: {{ .GetAddress }}
# Peer identity: /p2p/{{ .P2PKey | P2PKeyPublic}}

ShuttermintURL		= "{{ .ShuttermintURL }}"
# The layer 1 JSON RPC endpoitn
EthereumURL         = "{{ .EthereumURL }}"
# The JSON RPC endpoint where the contracts are accessible
ContractsURL               = "{{ .ContractsURL }}"
DeploymentDir       = "{{ .DeploymentDir }}"

# DatabaseURL looks like postgres://username:password@localhost:5432/database_name
# If it's empty, we use the standard PG* environment variables
DatabaseURL		= "{{ .DatabaseURL }}"
DKGPhaseLength		= {{ .DKGPhaseLength }}

# DKGStartBlockDelta is used to delay the start of the DKG process. The first block where the DKG
# process may start is the activation block - DKGStartBlockDelta
DKGStartBlockDelta   = {{ .DKGStartBlockDelta }}

# p2p configuration
ListenAddresses   = [{{ .ListenAddresses | QuoteList}}]
CustomBootstrapAddresses  = [{{ .CustomBootstrapAddresses | ToMultiAddrList | QuoteList}}]

ValidatorPublicKey	= "{{ .ValidatorPublicKey | printf "%x" }}"

# Secret Keys
EncryptionKey	= "{{ .EncryptionKey.ExportECDSA | FromECDSA | printf "%x" }}"
SigningKey	= "{{ .SigningKey | FromECDSA | printf "%x" }}"
P2PKey          = "{{ .P2PKey | P2PKey}}"

# HTTP interface
HTTPEnabled       = {{ .HTTPEnabled }}
HTTPListenAddress = "{{ .HTTPListenAddress }}"

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
	validatorPublicKey := make([]byte, ed25519.PublicKeySize)

	p2pkey, _, err := p2pcrypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return err
	}

	config.SigningKey = signingKey
	config.ValidatorPublicKey = validatorPublicKey
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
				medley.StringToEd25519PublicKey,
				medley.StringToEcdsaPrivateKey,
				medley.StringToEciesPrivateKey,
				medley.StringToAddress,
				medley.P2PKeyHook,
				medley.StringToEnvironment,
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
				medley.MultiaddrHook,
				medley.AddrInfoHook,
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

// GetAddress returns the keyper's Ethereum address.
func (config *Config) GetAddress() common.Address {
	return crypto.PubkeyToAddress(config.SigningKey.PublicKey)
}

func (config *Config) GetHTTPListenAddress() string {
	return config.HTTPListenAddress
}

func (config *Config) GetInstanceID() uint64 {
	return config.InstanceID
}

func (config *Config) GetDKGPhaseLength() *dkgphase.PhaseLength {
	return dkgphase.NewConstantPhaseLength(int64(config.DKGPhaseLength))
}

func (config *Config) GetValidatorPublicKey() ed25519.PublicKey {
	return config.ValidatorPublicKey
}

func (config *Config) GetEncryptionKey() *ecies.PrivateKey {
	return config.EncryptionKey
}

// WriteTOML writes a toml configuration file with the given config.
func (config *Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}
