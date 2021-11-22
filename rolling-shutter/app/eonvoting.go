package app

import "github.com/ethereum/go-ethereum/common"

// NewEonStartVoting creates a new EonStartVoting struct.
func NewEonStartVoting() *EonStartVoting {
	v := EonStartVoting{
		Voting:     NewVoting(),
		Candidates: []uint64{},
	}
	return &v
}

// AddVote adds or updates a vote for a certain activation block number.
func (v *EonStartVoting) AddVote(sender common.Address, activationBlockNumber uint64) {
	for i, b := range v.Candidates {
		if b == activationBlockNumber {
			v.AddVoteForIndex(sender, i)
			return
		}
	}

	v.Candidates = append(v.Candidates, activationBlockNumber)
	v.AddVoteForIndex(sender, len(v.Candidates)-1)
}

// Outcome checks if an activation block number has a majority and if so returns it.
func (v *EonStartVoting) Outcome(numRequiredVotes int) (uint64, bool) {
	i, success := v.OutcomeIndex(numRequiredVotes)
	if !success {
		return 0, false
	}
	return v.Candidates[i], true
}
