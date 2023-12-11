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

func NewBlockNumber() *BlockNumber {
	return &BlockNumber{
		Int: new(big.Int),
	}
}

type BlockNumber struct {
	*big.Int
}

func (k *BlockNumber) UnmarshalText(b []byte) error {
	k.Int = new(big.Int)
	if bytes.Equal(b, LatestStr) {
		k.Int.SetInt64(LatestBlockInt)
		return nil
	}
	return k.Int.UnmarshalText(b)
}

func (k *BlockNumber) IsLatest() bool {
	return k.Equal(LatestBlock)
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
