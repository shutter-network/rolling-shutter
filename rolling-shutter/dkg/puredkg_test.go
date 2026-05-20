package dkg

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"gotest.tools/v3/assert"
)

func TestReceiverIndicesForSender(t *testing.T) {
	cases := []struct {
		n      uint64
		sender uint64
		want   []uint64
	}{
		{n: 1, sender: 0, want: []uint64{}},
		{n: 3, sender: 0, want: []uint64{1, 2}},
		{n: 3, sender: 1, want: []uint64{0, 2}},
		{n: 4, sender: 3, want: []uint64{0, 1, 2}},
	}
	for _, tc := range cases {
		got := ReceiverIndicesForSender(tc.n, tc.sender)
		assert.Equal(t, len(got), len(tc.want))
		for i := range tc.want {
			assert.Equal(t, got[i], tc.want[i])
		}
	}
}

// TestPureDKGReplayMatchesLive verifies that applying the standard message
// set to a fresh PureDKG (the "replay" path) produces the same observable
// state as applying the same set to the same DKG that drove the messages
// (the "live" path). This is the property the rebuild-from-DB logic depends
// on for restart recovery.
func TestPureDKGReplayMatchesLive(t *testing.T) {
	const (
		eon        uint64 = 17
		numKeypers uint64 = 3
		threshold  uint64 = 2
	)

	live := make([]*puredkg.PureDKG, numKeypers)
	commitments := make([]puredkg.PolyCommitmentMsg, 0, numKeypers)
	allEvals := make([][]puredkg.PolyEvalMsg, numKeypers)
	for i := uint64(0); i < numKeypers; i++ {
		p := puredkg.NewPureDKG(eon, numKeypers, threshold, i)
		live[i] = &p
		commit, evals, err := p.StartPhase1Dealing()
		assert.NilError(t, err)
		commitments = append(commitments, commit)
		allEvals[i] = evals
	}

	for i := uint64(0); i < numKeypers; i++ {
		for j := uint64(0); j < numKeypers; j++ {
			if i == j {
				continue
			}
			assert.NilError(t, live[i].HandlePolyCommitmentMsg(commitments[j]))
		}
		for _, evalsFromJ := range allEvals {
			for _, ev := range evalsFromJ {
				if ev.Receiver != i {
					continue
				}
				assert.NilError(t, live[i].HandlePolyEvalMsg(ev))
			}
		}
	}

	replay := puredkg.NewPureDKG(eon, numKeypers, threshold, 0)
	for _, c := range commitments {
		assert.NilError(t, replay.HandlePolyCommitmentMsg(c))
	}
	for _, evalsFromJ := range allEvals {
		for _, ev := range evalsFromJ {
			if ev.Receiver != 0 {
				continue
			}
			assert.NilError(t, replay.HandlePolyEvalMsg(ev))
		}
	}

	for i := uint64(0); i < numKeypers; i++ {
		if i == replay.Keyper {
			continue
		}
		assert.Assert(t,
			liveCommitmentEqualsReplay(live[0].Commitments[i], replay.Commitments[i]),
			"commitment from %d mismatched between live and replay", i)
	}
	for i := uint64(0); i < numKeypers; i++ {
		if i == replay.Keyper {
			continue
		}
		liveEval := live[0].Evals[i]
		replayEval := replay.Evals[i]
		assert.Assert(t,
			(liveEval == nil) == (replayEval == nil),
			"eval from %d nil-state mismatched between live and replay", i,
		)
		if liveEval != nil {
			assert.Equal(t, 0, liveEval.Cmp(replayEval),
				"eval from %d differs between live and replay", i)
		}
	}
}

// TestPureDKGReplayLateAccusations replays only an accusation message at
// phase Off and asserts the puredkg accepts it. This is the second leg of
// the rebuild logic: an accusation stored from a chain event must be
// applicable to a freshly-built puredkg.
func TestPureDKGReplayLateAccusations(t *testing.T) {
	p := puredkg.NewPureDKG(1, 3, 2, 0)
	err := p.HandleAccusationMsg(puredkg.AccusationMsg{
		Eon: 1, Accuser: 1, Accused: 2,
	})
	assert.NilError(t, err)
}

// TestPureDKGReplayLateApologies likewise checks the apology path.
func TestPureDKGReplayLateApologies(t *testing.T) {
	p := puredkg.NewPureDKG(1, 3, 2, 0)
	evalRaw, _ := rand.Int(rand.Reader, big.NewInt(1<<60))
	err := p.HandleApologyMsg(puredkg.ApologyMsg{
		Eon: 1, Accuser: 0, Accused: 2, Eval: evalRaw,
	})
	assert.NilError(t, err)
}

// TestComputeResultAfterReplayWithSelfEval asserts the recovery path used
// by `startFinalizing`: rebuild puredkg from stored messages (commitments
// and per-receiver evals, including a self-eval row), set `pure.Phase` to
// Finalized directly, and call `ComputeResult`. The resulting eon public
// key must equal what a "live" non-restarted DKG would have produced.
func TestComputeResultAfterReplayWithSelfEval(t *testing.T) {
	const (
		eon        uint64 = 17
		numKeypers uint64 = 3
		threshold  uint64 = 2
	)

	// Drive a full live DKG: each keyper deals, cross-delivers messages,
	// and finalises in-process.
	live := make([]*puredkg.PureDKG, numKeypers)
	commitments := make([]puredkg.PolyCommitmentMsg, numKeypers)
	evalsBySender := make([][]puredkg.PolyEvalMsg, numKeypers)
	selfEvals := make([]*big.Int, numKeypers)
	for i := uint64(0); i < numKeypers; i++ {
		p := puredkg.NewPureDKG(eon, numKeypers, threshold, i)
		live[i] = &p
		commit, evals, err := p.StartPhase1Dealing()
		assert.NilError(t, err)
		commitments[i] = commit
		evalsBySender[i] = evals
		// puredkg consumes the self-eval inside StartPhase1Dealing; capture
		// what it set so the recovery path can replay it.
		selfEvals[i] = p.Evals[i]
	}
	for i := uint64(0); i < numKeypers; i++ {
		for j := uint64(0); j < numKeypers; j++ {
			if i == j {
				continue
			}
			assert.NilError(t, live[i].HandlePolyCommitmentMsg(commitments[j]))
		}
		for _, evals := range evalsBySender {
			for _, ev := range evals {
				if ev.Receiver != i {
					continue
				}
				assert.NilError(t, live[i].HandlePolyEvalMsg(ev))
			}
		}
		assert.NilError(t, live[i].HandlePolyCommitmentMsg(commitments[i]))
	}
	for i := uint64(0); i < numKeypers; i++ {
		assert.NilError(t, advanceLiveToFinalized(live[i]))
	}
	liveResult, err := live[0].ComputeResult()
	assert.NilError(t, err)

	// Now simulate the recovery path for keyper 0: start fresh, replay all
	// commitments, replay every received eval including the self-eval row,
	// then jump straight to Finalized and compute the result.
	replay := puredkg.NewPureDKG(eon, numKeypers, threshold, 0)
	for _, c := range commitments {
		assert.NilError(t, replay.HandlePolyCommitmentMsg(c))
	}
	for _, evals := range evalsBySender {
		for _, ev := range evals {
			if ev.Receiver != 0 {
				continue
			}
			assert.NilError(t, replay.HandlePolyEvalMsg(ev))
		}
	}
	// The self-eval (sender=0, receiver=0) is what `startDealing` now also
	// persists as a dedicated row. Replay it here.
	assert.NilError(t, replay.HandlePolyEvalMsg(puredkg.PolyEvalMsg{
		Eon: eon, Sender: 0, Receiver: 0, Eval: selfEvals[0],
	}))
	replay.Phase = puredkg.Finalized
	recoveredResult, err := replay.ComputeResult()
	assert.NilError(t, err)

	// The eon public key — which is what the success vote carries — must
	// match.
	livePK, err := liveResult.PublicKey.GobEncode()
	assert.NilError(t, err)
	recoveredPK, err := recoveredResult.PublicKey.GobEncode()
	assert.NilError(t, err)
	assert.Equal(t, string(livePK), string(recoveredPK))
}

func advanceLiveToFinalized(p *puredkg.PureDKG) error {
	p.StartPhase2Accusing()
	p.StartPhase3Apologizing()
	p.Finalize()
	return nil
}

func liveCommitmentEqualsReplay(a, b interface{}) bool {
	type marshaller interface {
		MarshalText() ([]byte, error)
	}
	am, aok := a.(marshaller)
	bm, bok := b.(marshaller)
	if !aok || !bok {
		return a == b
	}
	at, _ := am.MarshalText()
	bt, _ := bm.MarshalText()
	return string(at) == string(bt)
}
