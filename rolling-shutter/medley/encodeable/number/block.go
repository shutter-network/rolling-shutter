package number

import (
	"bytes"
	"errors"
	"math/big"
)

var (
	LatestBlockInt int64 = -1
	LatestBlock          = &BlockNumber{new(big.Int).SetInt64(LatestBlockInt)}
	LatestStr            = []byte("latest")
)

var ErrLatestBlockToUint = errors.New("'latest' block can't be converted to uint64")

func BigToBlockNumber(i *big.Int) *BlockNumber {
	if i == nil {
		return NewBlockNumber(nil)
	}
	return &BlockNumber{
		Int: i,
	}
}

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

func (k *BlockNumber) MarshalJSON() ([]byte, error) {
	return k.MarshalText()
}

func (k *BlockNumber) UnmarshalJSON(b []byte) error {
	return k.UnmarshalText(b)
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

func (k *BlockNumber) ToUInt64() (uint64, error) {
	p := k.ToUInt64Ptr()
	if p == nil {
		return 0, ErrLatestBlockToUint
	}
	return *p, nil
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
		return LatestStr, nil
	}
	return k.Int.MarshalText()
}
