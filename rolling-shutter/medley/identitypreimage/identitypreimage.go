package identitypreimage

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type IdentityPreimage []byte

// BytesToEpochID converts b to IdentityPreimage.
func BytesToIdentityPreimage(b []byte) IdentityPreimage {
	return IdentityPreimage(b)
}

// BigToIdentityPreimage converts n to an epoch id. It fails if n is too big.
func BigToIdentityPreimage(n *big.Int) (IdentityPreimage, error) {
	e := IdentityPreimage(n.Bytes())
	n2 := e.Big()
	if n2.Cmp(n) != 0 {
		return IdentityPreimage{}, errors.Errorf("input %d is too big to be an epoch id", n)
	}
	return e, nil
}

func HexToIdentityPreimage(n string) IdentityPreimage {
	return BytesToIdentityPreimage([]byte(n))
}

func Uint64ToIdentityPreimage(n uint64) IdentityPreimage {
	r, _ := BigToIdentityPreimage(new(big.Int).SetUint64(n))
	return r
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
	s := string(e)
	return s[2:6] + ".." + s[len(s)-4:]
}

func Equal(a, b IdentityPreimage) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}
