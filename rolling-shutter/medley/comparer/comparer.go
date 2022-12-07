package comparer

import (
	"bytes"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/crypto/ecies"
	gocmp "github.com/google/go-cmp/cmp"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
)

// P2PPrivKeyComparer is a gocmp comparer for use with gotest.tools/assert.DeepEqual.
var P2PPrivKeyComparer = gocmp.Comparer(func(k1, k2 p2pcrypto.PrivKey) bool {
	d1, _ := p2pcrypto.MarshalPrivateKey(k1)
	d2, _ := p2pcrypto.MarshalPrivateKey(k2)
	return bytes.Equal(d1, d2)
})

var EciesPublicKeyComparer = gocmp.Comparer(func(x, y *ecies.PublicKey) bool {
	return reflect.DeepEqual(x, y)
})

var EciesPrivateKeyComparer = gocmp.Comparer(func(x, y *ecies.PrivateKey) bool {
	return reflect.DeepEqual(x, y)
})

var (
	DurationComparer5MsDeviation = MakeDurationComparer(5 * time.Millisecond)
	DurationComparerStrict       = MakeDurationComparer(0 * time.Millisecond)
)

func MakeDurationComparer(allowedDeviation time.Duration) gocmp.Option {
	return gocmp.Comparer(
		func(a, b time.Duration) bool {
			var d int64
			aNs := a.Abs().Nanoseconds()
			bNs := b.Abs().Nanoseconds()
			if aNs <= bNs {
				d = bNs - aNs
			} else {
				d = aNs - bNs
			}
			return d <= allowedDeviation.Nanoseconds()
		})
}
