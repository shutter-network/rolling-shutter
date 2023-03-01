package config

import (
	"crypto/ecdsa"
	"io"
	"text/template"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/p2p"
)

type Config struct {
	ListenAddresses          []multiaddr.Multiaddr
	CustomBootstrapAddresses []peer.AddrInfo
	Environment              p2p.Environment
	EthereumURL              string
	SequencerURL             string
	ContractsURL             string
	DeploymentDir            string
	DatabaseURL              string
	HTTPListenAddress        string

	EthereumKey *ecdsa.PrivateKey
	P2PKey      p2pcrypto.PrivKey

	InstanceID uint64

	EpochDuration                time.Duration
	ExecutionBlockDelay          uint32
	BatchIndexAcceptenceInterval uint32
}

var configTemplate = `# Shutter collator config
# Ethereum address: {{ .EthereumAddress }}
# Peer identity: /p2p/{{ .P2PKey | P2PKeyPublic}}

# L1 node URL
EthereumURL     = "{{ .EthereumURL }}"
DeploymentDir   = "{{ .DeploymentDir }}"

# DatabaseURL looks like postgres://username:password@localhost:5432/database_name
# If it's empty, we use the standard PG* environment variables
DatabaseURL     = "{{ .DatabaseURL }}"

HTTPListenAddress = "{{ .HTTPListenAddress }}"

# JSON RPC endpoint of the sequencer to which batches will be submitted
SequencerURL = "{{ .SequencerURL }}"

# JSON RPC endpoint of the node where the contracts are deployed
ContractsURL = "{{ .ContractsURL }}"

# p2p configuration
ListenAddresses   = [{{ .ListenAddresses | QuoteList}}]
CustomBootstrapAddresses  = [{{ .CustomBootstrapAddresses | ToMultiAddrList | QuoteList}}]

# Secret Keys
EthereumKey     = "{{ .EthereumKey | FromECDSA | printf "%x" }}"
P2PKey          = "{{ .P2PKey | P2PKey}}"

# ID shared by all shutter participants for common instance
InstanceID = {{ .InstanceID }}

# The duration of an epoch
EpochDuration = "{{ .EpochDuration }}"

# Number of blocks to backdate batches
ExecutionBlockDelay = {{ .ExecutionBlockDelay }}
`

var tmpl *template.Template = medley.MustBuildTemplate("collator", configTemplate)

// WriteTOML writes a toml configuration file with the given config.
func (config *Config) WriteTOML(w io.Writer) error {
	return tmpl.Execute(w, config)
}

// Unmarshal unmarshals a collator.Config from the given Viper object.
func (config *Config) Unmarshal(v *viper.Viper) error {
	err := v.Unmarshal(
		config,
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				medley.MultiaddrHook,
				medley.AddrInfoHook,
				medley.P2PKeyHook,
				medley.StringToEcdsaPrivateKey,
				medley.StringToEnvironment,
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
