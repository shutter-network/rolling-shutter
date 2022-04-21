package app

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"gotest.tools/v3/assert"

	"github.com/shutter-network/shutter/shlib/shtest"
)

func TestEonStartVoting(t *testing.T) {
	v := NewEonStartVoting()
	_, s := v.Outcome(2)
	assert.Assert(t, !s)

	a1 := common.BigToAddress(big.NewInt(0))
	a2 := common.BigToAddress(big.NewInt(1))
	a3 := common.BigToAddress(big.NewInt(3))
	e1 := uint64(3)
	e2 := uint64(5)

	v.SetVote(a1, e1)
	_, s = v.Outcome(2)
	assert.Assert(t, !s)

	v.SetVote(a2, e1)
	r, s := v.Outcome(2)
	assert.Assert(t, s)
	assert.Equal(t, e1, r)

	v.SetVote(a2, e2)
	_, s = v.Outcome(2)
	assert.Assert(t, !s)

	v.SetVote(a3, e2)
	r, s = v.Outcome(2)
	assert.Assert(t, s)
	assert.Equal(t, e2, r)
	dst := NewEonStartVoting()
	shtest.EnsureGobable(t, v, dst)
}
