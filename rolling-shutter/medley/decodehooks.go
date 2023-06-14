package medley

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"encoding"
	"encoding/hex"
	"fmt"
	"net/url"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mitchellh/mapstructure"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/env"
)

// MultiaddrHook is a mapstructure decode hook for multiaddrs.
func MultiaddrHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	if t != reflect.TypeOf((*multiaddr.Multiaddr)(nil)).Elem() {
		return data, nil
	}

	return multiaddr.NewMultiaddr(data.(string))
}

func AddrInfoHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	if t != reflect.TypeOf((*peer.AddrInfo)(nil)).Elem() {
		return data, nil
	}

	addrInfo, err := peer.AddrInfoFromString(data.(string))
	if err != nil {
		return nil, err
	}

	return addrInfo, nil
}

func P2PKeyHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	var privkey p2pcrypto.PrivKey

	if f.Kind() != reflect.String || t != reflect.TypeOf(&privkey).Elem() {
		return data, nil
	}

	k, err := p2pcrypto.ConfigDecodeKey(data.(string))
	if err != nil {
		return nil, err
	}
	privkey, err = p2pcrypto.UnmarshalPrivateKey(k)
	if err != nil {
		return nil, err
	}
	return privkey, nil
}

func StringToURL(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(&url.URL{}) {
		return data, nil
	}
	return url.Parse(data.(string))
}

func TextUnmarshalerHook(from reflect.Value, to reflect.Value) (interface{}, error) {
	data := from.Interface()
	if from.Kind() != reflect.String {
		return data, nil
	}
	// create new instance
	toNew := reflect.New(to.Type())
	resultPtr := toNew.Interface()
	umshl, ok := resultPtr.(encoding.TextUnmarshaler)
	if !ok {
		return data, nil
	}
	err := umshl.UnmarshalText([]byte(data.(string)))
	if err != nil {
		return nil, err
	}
	if to.Kind() == reflect.Pointer {
		return umshl, nil
	}
	// if to type is no ptr type,
	// return the element
	return toNew.Elem().Interface(), nil
}

func TextMarshalerHook(from reflect.Value, to reflect.Value) (interface{}, error) {
	if from.Kind() == reflect.Ptr {
		from = from.Elem()
	}

	data := from.Interface()
	if to.Kind() != reflect.String {
		return data, nil
	}
	fromType := from.Type()
	result := reflect.New(fromType).Interface()
	_, ok := result.(encoding.TextMarshaler)
	if !ok {
		return data, nil
	}

	marshaller, ok := data.(encoding.TextMarshaler)
	if !ok {
		return data, nil
	}
	mshl, err := marshaller.MarshalText()
	if err != nil {
		return nil, err
	}
	return string(mshl), nil
}

func StringToEd25519PublicKey(
	f reflect.Type,
	t reflect.Type,
	data interface{},
) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(ed25519.PublicKey{}) {
		return data, nil
	}
	p, err := hex.DecodeString(data.(string))
	if err != nil {
		return nil, err
	}
	if len(p) != ed25519.PublicKeySize {
		return nil, errors.New("badly formed ed25519 public key")
	}
	return ed25519.PublicKey(p), nil
}

func StringToEd25519PrivateKey(
	f reflect.Type,
	t reflect.Type,
	data interface{},
) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(ed25519.PrivateKey{}) {
		return data, nil
	}
	seed, err := hex.DecodeString(data.(string))
	if err != nil {
		return nil, err
	}
	if len(seed) != ed25519.SeedSize {
		return nil, errors.Errorf(
			"invalid seed length %d (must be %d)",
			len(seed),
			ed25519.SeedSize,
		)
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

func StringToEnvironment(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(env.EnvironmentStaging) {
		return data, nil
	}
	result, err := env.ParseEnvironment(data.(string))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func StringToEcdsaPrivateKey(
	f reflect.Type,
	t reflect.Type,
	data interface{},
) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(&ecdsa.PrivateKey{}) {
		return data, nil
	}
	return crypto.HexToECDSA(data.(string))
}

func StringToEciesPrivateKey(
	f reflect.Type,
	t reflect.Type,
	data interface{},
) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(&ecies.PrivateKey{}) {
		return data, nil
	}
	encryptionKeyECDSA, err := crypto.HexToECDSA(data.(string))
	if err != nil {
		return nil, err
	}

	return ecies.ImportECDSA(encryptionKeyECDSA), nil
}

func StringToAddress(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
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

func mapstructureDecode(input, result any, hookFunc mapstructure.DecodeHookFunc) error {
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			Result:     result,
			DecodeHook: hookFunc,
		})
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

func MapstructureMarshal(input, result any) error {
	return mapstructureDecode(
		input,
		result,
		mapstructure.ComposeDecodeHookFunc(
			TextMarshalerHook,
		),
	)
}

func MapstructureUnmarshal(input, result any) error {
	return mapstructureDecode(
		input,
		result,
		mapstructure.ComposeDecodeHookFunc(
			TextUnmarshalerHook,
		),
	)
}
