package medley

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
	multiaddr "github.com/multiformats/go-multiaddr"
)

// MultiaddrHook returns a mapstructure decode hook for multiaddrs.
func MultiaddrHook() mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf((*multiaddr.Multiaddr)(nil)).Elem() {
			return data, nil
		}

		return multiaddr.NewMultiaddr(data.(string))
	}
}
