package number

import (
	"bytes"
	"math/big"
)

var (
	LatestBlockInt int64 = -1
	LatestBlock          = &BlockNumber{new(big.Int).SetInt64(LatestBlockInt)}
	LatestStr            = []byte("latest")
)

func NewBlockNumber(u *uint64) *BlockNumber {
	b := &BlockNumber{
		Int: &big.Int{},
	}
	if u == nil {
		b.SetInt64(LatestBlockInt)
	} else {
		b.SetUint64(*u)
	}
	return b
}

type BlockNumber struct {
	*big.Int
}

func (k *BlockNumber) UnmarshalText(b []byte) error {
	k.Int = &big.Int{}
	if bytes.Equal(b, LatestStr) {
		k.Int.SetInt64(LatestBlockInt)
		return nil
	}
	return k.Int.UnmarshalText(b)
}

func (k *BlockNumber) IsLatest() bool {
	return k.Equal(LatestBlock)
}

func (k *BlockNumber) ToUInt64Ptr() *uint64 {
	if k.IsLatest() {
		return nil
	}
	u := k.Uint64()
	return &u
}

func (k *BlockNumber) Equal(b *BlockNumber) bool {
	return k.Int.Cmp(b.Int) == 0
}

func (k *BlockNumber) MarshalText() ([]byte, error) {
	if k.IsLatest() {
		return []byte("latest"), nil
	}
	return k.Int.MarshalText()
}
