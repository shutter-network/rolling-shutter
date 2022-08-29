package epochid

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type EpochID common.Hash

// BytesToEpochID converts b to an epoch id. It fails if b is not 32 bytes.
func BytesToEpochID(b []byte) (EpochID, error) {
	if len(b) != len(common.Hash{}) {
		return EpochID{}, errors.Errorf("epoch id must be %d bytes, got %d", len(common.Hash{}), len(b))
	}
	return EpochID(common.BytesToHash(b)), nil
}

// BigToEpochID converts n to an epoch id. It fails if n is too big.
func BigToEpochID(n *big.Int) (EpochID, error) {
	e := EpochID(common.BigToHash(n))
	n2 := e.Big()
	if n2.Cmp(n) != 0 {
		return EpochID{}, errors.Errorf("input %d is too big to be an epoch id", n)
	}
	return e, nil
}

func HexToEpochID(n string) (EpochID, error) {
	return BytesToEpochID(common.FromHex(n))
}

func Uint64ToEpochID(n uint64) (EpochID, error) {
	return BigToEpochID(new(big.Int).SetUint64(n))
}

func (e EpochID) Bytes() []byte {
	return common.Hash(e).Bytes()
}

func (e EpochID) Big() *big.Int {
	return common.Hash(e).Big()
}

func (e EpochID) Uint64() uint64 {
	return e.Big().Uint64()
}

func (e EpochID) Hex() string {
	return common.Hash(e).Hex()
}

func (e EpochID) String() string {
	s := common.Hash(e).String()
	return s[2:6] + ".." + s[len(s)-4:]
}

func Equal(a, b EpochID) bool {
	return bytes.Equal(a.Bytes(), b.Bytes())
}
