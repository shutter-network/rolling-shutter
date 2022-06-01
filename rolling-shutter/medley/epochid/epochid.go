package epochid

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type EpochID common.Hash

// BytesToEpochID converts b to an epoch id. It fails if b is not 32 bytes.
func BytesToEpochID(b []byte) (EpochID, error) {
	if len(b) != len(common.Hash{}) {
		return EpochID{}, errors.Errorf("epoch id must be %d bytes, got %d", len(b), len(common.Hash{}))
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

func (e EpochID) Bytes() []byte {
	return common.Hash(e).Bytes()
}

func (e EpochID) Big() *big.Int {
	return common.Hash(e).Big()
}

func (e EpochID) String() string {
	s := e.String()
	return s[2:6] + ".." + s[len(s)-4:]
}
