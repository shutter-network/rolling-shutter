package medley

import (
	"encoding/hex"
	"reflect"

	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	multiaddr "github.com/multiformats/go-multiaddr"

	"github.com/shutter-network/shutter/shlib/shcrypto/shbls"
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

func BLSSecretKeyHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	if t != reflect.TypeOf((*shbls.SecretKey)(nil)).Elem() {
		return data, nil
	}

	b, err := hex.DecodeString(data.(string))
	if err != nil {
		return nil, err
	}

	key := new(shbls.SecretKey)
	if err := key.Unmarshal(b); err != nil {
		return nil, err
	}
	return key, nil
}

func BLSPublicKeyHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	if t != reflect.TypeOf((*shbls.PublicKey)(nil)).Elem() {
		return data, nil
	}

	b, err := hex.DecodeString(data.(string))
	if err != nil {
		return nil, err
	}

	key := new(shbls.PublicKey)
	if err := key.Unmarshal(b); err != nil {
		return nil, err
	}
	return key, nil
}
