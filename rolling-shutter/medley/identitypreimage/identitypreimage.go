package identitypreimage

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type IdentityPreimage []byte

// BigToIdentityPreimage converts n to an epoch id. It fails if n is too big.
func BigToIdentityPreimage(n *big.Int) IdentityPreimage {
	return IdentityPreimage(n.Bytes())
}

func HexToIdentityPreimage(n string) IdentityPreimage {
	return IdentityPreimage(common.FromHex(n))
}

func Uint64ToIdentityPreimage(n uint64) IdentityPreimage {
	return BigToIdentityPreimage(new(big.Int).SetUint64(n))
}

func (e IdentityPreimage) Bytes() []byte {
	return []byte(e)
}

func (e IdentityPreimage) Big() *big.Int {
	return new(big.Int).SetBytes(e)
}

func (e IdentityPreimage) Uint64() uint64 {
	return e.Big().Uint64()
}

func (e IdentityPreimage) Hex() string {
	return common.Bytes2Hex([]byte(e))
}

func (e IdentityPreimage) String() string {
	var hash common.Hash
	hash.SetBytes(e.Bytes())
	s := hash.String()
	return s[2:6] + ".." + s[len(s)-4:]
}

func Equal(a, b IdentityPreimage) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}
