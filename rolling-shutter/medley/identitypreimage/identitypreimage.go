package identitypreimage

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type IdentityPreimage []byte

func BigToIdentityPreimage(n *big.Int) IdentityPreimage {
	return IdentityPreimage(n.Bytes())
}

func HexToIdentityPreimage(n string) (IdentityPreimage, error) {
	b, err := hexutil.Decode(n)
	return IdentityPreimage(b), err
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
	return hexutil.Encode(e)
}

func (e IdentityPreimage) String() string {
	s := e.Hex()
	if len(s) < 10 {
		return s[2:]
	}
	return s[2:6] + ".." + s[len(s)-4:]
}

func Equal(a, b IdentityPreimage) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}

func (e IdentityPreimage) MarshalText() ([]byte, error) { //nolint:unparam
	return []byte(e.Hex()), nil
}

func (e *IdentityPreimage) UnmarshalText(input []byte) error {
	val, err := HexToIdentityPreimage(string(input))
	*e = val.Bytes()
	return err
}
