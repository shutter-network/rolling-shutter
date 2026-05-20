package txsender

import (
	"math/big"
	"testing"

	"github.com/jackc/pgtype"
	"gotest.tools/assert"
)

func TestBigIntNumericRoundtrip(t *testing.T) {
	cases := []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(1_000_000_000_000_000_000), // 1 ETH in wei
		new(big.Int).Mul(big.NewInt(1e18), big.NewInt(1e9)),
	}
	for _, want := range cases {
		num, err := bigIntToNumeric(want)
		assert.NilError(t, err)
		got, err := numericToBigInt(num)
		assert.NilError(t, err)
		assert.Equal(t, want.Cmp(got), 0, "roundtrip mismatch: want %s got %s", want, got)
	}
}

func TestNilBigIntEncodesAsZero(t *testing.T) {
	num, err := bigIntToNumeric(nil)
	assert.NilError(t, err)
	got, err := numericToBigInt(num)
	assert.NilError(t, err)
	assert.Equal(t, got.Sign(), 0)
}

func TestNullNumericDecodesAsZero(t *testing.T) {
	got, err := numericToBigInt(pgtype.Numeric{Status: pgtype.Null})
	assert.NilError(t, err)
	assert.Equal(t, got.Sign(), 0)
}
