package medley

import (
	"reflect"

	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	multiaddr "github.com/multiformats/go-multiaddr"
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
