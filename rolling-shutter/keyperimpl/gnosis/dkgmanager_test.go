package gnosis

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/shutter-network/shutter/shlib/puredkg"

	"gotest.tools/v3/assert"
)

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

	// Generate three live PureDKG instances and drive them through Phase 1.
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

	// Cross-deliver: each live keyper consumes the others' messages.
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

	// Build a fresh "replay" PureDKG for keyper 0 and feed in the same
	// messages while staying at phase Off (the rebuild path leaves the
	// in-memory phase alone — live phase transitions are driven externally).
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
	// Provide a non-trivial big.Int so puredkg's validity check accepts it.
	evalRaw, _ := rand.Int(rand.Reader, big.NewInt(1<<60))
	err := p.HandleApologyMsg(puredkg.ApologyMsg{
		Eon: 1, Accuser: 0, Accused: 2, Eval: evalRaw,
	})
	assert.NilError(t, err)
}

func liveCommitmentEqualsReplay(a, b interface{}) bool {
	// Both should be `*shcrypto.Gammas`; compare by their textual form which
	// is the canonical serialisation. Avoid importing shcrypto here to keep
	// the test focused on observable behaviour.
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
