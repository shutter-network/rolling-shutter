package keyper

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/mitchellh/mapstructure"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/shutter-network/shutter/shuttermint/medley"
)

// Config contains validated configuration parameters for the keyper client.
type Config struct {
	ShuttermintURL string
	DBDir          string
	DatabaseURL    string
	SigningKey     *ecdsa.PrivateKey
	ValidatorKey   ed25519.PrivateKey `mapstructure:"ValidatorSeed"`
	EncryptionKey  *ecies.PrivateKey
	P2PKey         p2pcrypto.PrivKey
	DKGPhaseLength uint64 // in shuttermint blocks
	ListenAddress  multiaddr.Multiaddr
	PeerMultiaddrs []multiaddr.Multiaddr
}

const configTemplate = `# Shutter keyper configuration for {{ .Address }}, /p2p/{{ .P2PKey | P2PKeyPublic}}

ShuttermintURL		= "{{ .ShuttermintURL }}"
DBDir			= "{{ .DBDir }}"

# DatabaseURL looks like postgres://username:password@localhost:5432/database_name
# It it's empty, we use the standard PG* environment variables
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
`

var tmpl *template.Template

func p2pKeyPublic(privkey p2pcrypto.PrivKey) string {
	id, _ := peer.IDFromPublicKey(privkey.GetPublic())
	return id.Pretty()
}

func p2pKey(privkey p2pcrypto.PrivKey) string {
	d, _ := p2pcrypto.MarshalPrivateKey(privkey)
	return p2pcrypto.ConfigEncodeKey(d)
}

func QuoteList(lst []multiaddr.Multiaddr) string {
	var strlist []string
	for _, x := range lst {
		// We use json.Marshal here, not sure if it's the right thing to do, since we're
		// writing TOML
		d, _ := json.Marshal(x.String())
		strlist = append(strlist, string(d))
	}

	return strings.Join(strlist, ", ")
}

func init() {
	var err error
	tmpl, err = template.New("keyper").Funcs(template.FuncMap{
		"FromECDSA":    crypto.FromECDSA,
		"QuoteList":    QuoteList,
		"P2PKey":       p2pKey,
		"P2PKeyPublic": p2pKeyPublic,
	}).Parse(configTemplate)
	if err != nil {
		panic(err)
	}
}

func stringToEd25519PrivateKey(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(ed25519.PrivateKey{}) {
		return data, nil
	}
	seed, err := hex.DecodeString(data.(string))
	if err != nil {
		return nil, err
	}
	if len(seed) != ed25519.SeedSize {
		return nil, errors.Errorf("invalid seed length %d (must be %d)", len(seed), ed25519.SeedSize)
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

func stringToEcdsaPrivateKey(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(&ecdsa.PrivateKey{}) {
		return data, nil
	}
	return crypto.HexToECDSA(data.(string))
}

func stringToEciesPrivateKey(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(&ecies.PrivateKey{}) {
		return data, nil
	}
	encryptionKeyECDSA, err := crypto.HexToECDSA(data.(string))
	if err != nil {
		return nil, err
	}

	return ecies.ImportECDSA(encryptionKeyECDSA), nil
}

func stringToAddress(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(common.Address{}) {
		return data, nil
	}

	ds := data.(string)
	addr := common.HexToAddress(ds)
	if addr.Hex() != ds {
		return nil, fmt.Errorf("not a checksummed address: %s", ds)
	}
	return addr, nil
}

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
				stringToEd25519PrivateKey,
				stringToEcdsaPrivateKey,
				stringToEciesPrivateKey,
				stringToAddress,
				medley.P2PKeyHook,
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
				medley.MultiaddrHook(),
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
