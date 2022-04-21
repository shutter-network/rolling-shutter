package app

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var errAlreadyVoted = errors.New("sender already voted")

// Equals is used to parametrize equality comparison in the Voting type. We need this because we
// cannot use the default comparison operator, because BatchConfig objects are not comparable. Also
// we cannot define a custom comparison method on the type, because we want to be able to also vote
// on basic types like uint64. This is being used in SetVote below, by using the Equals method on
// the null value of the given Equals type.
type Equals[T any] interface {
	Equals(a, b T) bool
}

// Voting is used to store votes. Each ethereum address can vote on one 'candidate'.
type Voting[T any, E Equals[T]] struct {
	Votes      map[common.Address]int
	Candidates []T
}

// NewVoting creates an empty Voting struct, where ethereum addresses can vote on candidates of
// type T.
func NewVoting[T any, E Equals[T]]() Voting[T, E] {
	return Voting[T, E]{
		Votes:      make(map[common.Address]int),
		Candidates: nil,
	}
}

// SetVote registers the given vote of the given address. Only the last vote of a given ethereum
// address is stored.
func (v *Voting[T, E]) SetVote(sender common.Address, candidate T) {
	var eq E
	for i, c := range v.Candidates {
		if eq.Equals(candidate, c) {
			v.Votes[sender] = i
			return
		}
	}
	v.Candidates = append(v.Candidates, candidate)
	v.Votes[sender] = len(v.Candidates) - 1
}

// AddVote registers the given vote of the given address. This function returns an error if the
// address already voted.
func (v *Voting[T, _]) AddVote(sender common.Address, candidate T) error {
	_, ok := v.Votes[sender]
	if ok {
		return errAlreadyVoted
	}
	v.SetVote(sender, candidate)
	return nil
}

// outcomeIndex checks if one of the candidate indices has more than numRequiredVotes.
func (v *Voting[_, _]) outcomeIndex(numRequiredVotes int) (int, bool) {
	numVotes := make(map[int]int)

	for _, vote := range v.Votes {
		numVotes[vote]++
	}
	for index, votes := range numVotes {
		if votes >= numRequiredVotes {
			return index, true
		}
	}
	return -1, false
}

// Outcome checks if one of the votes has received at least numRequiredVotes votes and returns it.
func (v *Voting[T, _]) Outcome(numRequiredVotes int) (T, bool) {
	idx, ok := v.outcomeIndex(numRequiredVotes)
	if !ok {
		var n T
		return n, false
	}
	return v.Candidates[idx], true
}
