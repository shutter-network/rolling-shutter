package dkg

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// ReceiverIndicesForSender returns the keyper indices that should be receivers
// of polyEvals submitted by `senderIndex` within a keyper set of size `n`:
// all indices in [0, n) except `senderIndex`, in ascending order.
func ReceiverIndicesForSender(n, senderIndex uint64) []uint64 {
	out := make([]uint64, 0, n-1)
	for i := uint64(0); i < n; i++ {
		if i == senderIndex {
			continue
		}
		out = append(out, i)
	}
	return out
}

// buildPureDKG reconstructs the in-memory puredkg state for a single DKG
// attempt at the phase the corresponding maybe-function expects to operate
// in. The DB is the source of truth; nothing is cached between invocations.
//
// The caller is responsible for the keyper-set lookup and membership check
// (both run as pool queries inside `processDKG` before this function is
// entered). `keypers`, `ownIndex`, and `threshold` are passed in as
// parameters so this function performs no observer-db reads.
//
// Returns (nil, nil) when this keyper has no `dkg_initial_states` row for a
// non-Dealing phase (Keyper crashed before dealing — they cannot
// accuse/apologize/finalize via local replay because the polynomial is
// unrecoverable).
//
// Phase-by-phase reconstruction:
//
//   - PhaseDealing: fresh PureDKG (Phase=Off). Apply already-received
//     PolyCommitment + PolyEval messages. The caller (`maybeDeal`) decides
//     whether to call StartPhase1Dealing based on the dkg_initial_states
//     idempotency check.
//   - PhaseAccusing: load from dkg_initial_states (Phase=Dealing, polynomial
//     set, self-eval populated). Apply PolyCommitment + PolyEval messages.
//     The caller (`maybeAccuse`) calls StartPhase2Accusing.
//   - PhaseApologizing: load from dkg_initial_states. Apply PolyCommitment +
//     PolyEval. Fast-forward via StartPhase2Accusing (output discarded —
//     accusations were already committed to chain). Apply Accusation
//     messages. The caller (`maybeApologize`) calls StartPhase3Apologizing.
//   - PhaseFinalizing: load from dkg_initial_states. Apply PolyCommitment +
//     PolyEval. Fast-forward via StartPhase2Accusing + StartPhase3Apologizing
//     (polynomial alive from the initial state makes Phase3 a no-op for
//     accusations not addressed to us). Apply Apology messages. The caller
//     sets Phase=Finalized directly and calls ComputeResult.
//
// Messages are applied in phase order: each `Start*` call is interleaved
// before the next message type so puredkg's per-handler phase guards are
// satisfied (`HandleAccusationMsg` requires Phase ≤ Accusing,
// `HandleApologyMsg` requires Phase ≤ Apologizing).
func (m *Manager) buildPureDKG(
	ctx context.Context,
	tx pgx.Tx,
	keyperSetIndex, retryCounter int64,
	blockPhase Phase,
	keypers []common.Address,
	ownIndex uint64,
	threshold uint64,
) (*puredkg.PureDKG, error) {
	queries := corekeyperdb.New(tx)

	var pure *puredkg.PureDKG
	switch blockPhase {
	case PhaseDealing:
		p := puredkg.NewPureDKG(
			uint64(keyperSetIndex), //nolint:gosec // G115: keyper set index is bounded by the on-chain contract
			uint64(len(keypers)),
			threshold,
			ownIndex,
		)
		pure = &p
	case PhaseAccusing, PhaseApologizing, PhaseFinalizing:
		row, err := queries.GetDKGInitialState(ctx, corekeyperdb.GetDKGInitialStateParams{
			KeyperSetIndex: keyperSetIndex,
			RetryCounter:   retryCounter,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, errors.Wrap(err, "load dkg initial state")
		}
		pure, err = shdb.DecodePureDKG(row.PuredkgBytes)
		if err != nil {
			return nil, errors.Wrap(err, "decode dkg initial state")
		}
		// gob converts nil pointers in `[]*shcrypto.Gammas` / `[]*big.Int`
		// slices into non-nil zero-value pointers on decode (via the
		// elements' GobDecoder), which would defeat puredkg's
		// "duplicate msg" check during the subsequent DB replay. Reset
		// the slices, then re-derive the self-eval and self-commitment
		// from the loaded polynomial — both are fully determined by the
		// polynomial, so we do not depend on chain-syncer indexing of
		// our own submitDealing event to populate these slots.
		pure.Commitments = make([]*shcrypto.Gammas, pure.NumKeypers)
		pure.Evals = make([]*big.Int, pure.NumKeypers)
		//nolint:gosec // G115: keyper index is bounded by the keyper set size
		pure.Evals[pure.Keyper] = pure.Polynomial.EvalForKeyper(int(pure.Keyper))
		pure.Commitments[pure.Keyper] = pure.Polynomial.Gammas()
	case PhaseNone:
		return nil, nil
	default:
		return nil, nil
	}

	if err := m.replayCommitmentsAndEvals(ctx, queries, pure, ownIndex, keyperSetIndex, retryCounter); err != nil {
		return nil, err
	}

	if blockPhase == PhaseDealing || blockPhase == PhaseAccusing {
		return pure, nil
	}

	// Fast-forward into Accusing so HandleAccusationMsg accepts the stored
	// rows. The output is discarded — the on-chain accusations are the
	// authoritative copy.
	_ = pure.StartPhase2Accusing()
	if err := m.replayAccusations(ctx, queries, pure, keyperSetIndex, retryCounter); err != nil {
		return nil, err
	}

	if blockPhase == PhaseApologizing {
		return pure, nil
	}

	// PhaseFinalizing: fast-forward into Apologizing and replay apologies.
	// StartPhase3Apologizing reads pure.Polynomial — which is alive because
	// it was loaded from dkg_initial_states.
	_ = pure.StartPhase3Apologizing()
	if err := m.replayApologies(ctx, queries, pure, keyperSetIndex, retryCounter); err != nil {
		return nil, err
	}
	return pure, nil
}

// replayCommitmentsAndEvals feeds stored PolyCommitment and PolyEval rows
// back into `pure`. Both handlers require Phase ≤ Dealing; the caller is
// responsible for the puredkg being at that phase or below.
//
// Duplicate-row errors (e.g. our own PolyCommitment row indexed by the
// chain syncer after the on-chain submitDealing event, which collides
// with `pure.Commitments[ownIndex]` already derived from the polynomial)
// are logged at debug level and ignored.
func (m *Manager) replayCommitmentsAndEvals(
	ctx context.Context,
	queries *corekeyperdb.Queries,
	pure *puredkg.PureDKG,
	ownIndex uint64,
	keyperSetIndex, retryCounter int64,
) error {
	eonForMsg := uint64(keyperSetIndex) //nolint:gosec // G115: keyper set index is bounded by the on-chain contract

	commitments, err := queries.GetDKGPolyCommitments(ctx, corekeyperdb.GetDKGPolyCommitmentsParams{
		KeyperSetIndex: keyperSetIndex,
		RetryCounter:   retryCounter,
	})
	if err != nil {
		return errors.Wrap(err, "load stored poly commitments")
	}
	for _, c := range commitments {
		gammas := &shcrypto.Gammas{}
		if err := gammas.Unmarshal(c.Commitment); err != nil {
			return errors.Wrapf(err, "decode commitment from sender %d", c.KeyperIndex)
		}
		err := pure.HandlePolyCommitmentMsg(puredkg.PolyCommitmentMsg{
			Eon:    eonForMsg,
			Sender: uint64(c.KeyperIndex), //nolint:gosec // G115: keyper index is bounded by the keyper set size
			Gammas: gammas,
		})
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-set-index", keyperSetIndex).
				Int64("retry-counter", retryCounter).
				Int64("sender", c.KeyperIndex).
				Msg("ignoring stored commitment on replay")
		}
	}

	polyEvals, err := queries.GetDKGPolyEvals(ctx, corekeyperdb.GetDKGPolyEvalsParams{
		KeyperSetIndex: keyperSetIndex,
		RetryCounter:   retryCounter,
	})
	if err != nil {
		return errors.Wrap(err, "load stored poly evals")
	}
	for _, ev := range polyEvals {
		if uint64(ev.ReceiverIndex) != ownIndex { //nolint:gosec // G115: keyper index is bounded by the keyper set size
			continue
		}
		eval, err := m.decryptPolyEval(ev.EncryptedEval)
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-set-index", keyperSetIndex).
				Int64("retry-counter", retryCounter).
				Int64("sender", ev.SenderIndex).
				Msg("ignoring undecryptable poly eval on replay")
			continue
		}
		err = pure.HandlePolyEvalMsg(puredkg.PolyEvalMsg{
			Eon:      eonForMsg,
			Sender:   uint64(ev.SenderIndex), //nolint:gosec // G115: keyper index is bounded by the keyper set size
			Receiver: ownIndex,
			Eval:     eval,
		})
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-set-index", keyperSetIndex).
				Int64("retry-counter", retryCounter).
				Int64("sender", ev.SenderIndex).
				Msg("ignoring poly eval on replay")
		}
	}
	return nil
}

func (m *Manager) replayAccusations(
	ctx context.Context,
	queries *corekeyperdb.Queries,
	pure *puredkg.PureDKG,
	keyperSetIndex, retryCounter int64,
) error {
	eonForMsg := uint64(keyperSetIndex) //nolint:gosec // G115: keyper set index is bounded by the on-chain contract
	accusations, err := queries.GetDKGAccusations(ctx, corekeyperdb.GetDKGAccusationsParams{
		KeyperSetIndex: keyperSetIndex,
		RetryCounter:   retryCounter,
	})
	if err != nil {
		return errors.Wrap(err, "load stored accusations")
	}
	for _, a := range accusations {
		err := pure.HandleAccusationMsg(puredkg.AccusationMsg{
			Eon:     eonForMsg,
			Accuser: uint64(a.AccuserIndex), //nolint:gosec // G115: keyper index is bounded by the keyper set size
			Accused: uint64(a.AccusedIndex), //nolint:gosec // G115: keyper index is bounded by the keyper set size
		})
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-set-index", keyperSetIndex).
				Int64("retry-counter", retryCounter).
				Int64("accuser", a.AccuserIndex).
				Int64("accused", a.AccusedIndex).
				Msg("ignoring accusation on replay")
		}
	}
	return nil
}

func (m *Manager) replayApologies(
	ctx context.Context,
	queries *corekeyperdb.Queries,
	pure *puredkg.PureDKG,
	keyperSetIndex, retryCounter int64,
) error {
	eonForMsg := uint64(keyperSetIndex) //nolint:gosec // G115: keyper set index is bounded by the on-chain contract
	apologies, err := queries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperSetIndex: keyperSetIndex,
		RetryCounter:   retryCounter,
	})
	if err != nil {
		return errors.Wrap(err, "load stored apologies")
	}
	for _, ap := range apologies {
		eval := new(big.Int).SetBytes(ap.PolyEval)
		err := pure.HandleApologyMsg(puredkg.ApologyMsg{
			Eon:     eonForMsg,
			Accuser: uint64(ap.AccuserIndex),    //nolint:gosec // G115: keyper index is bounded by the keyper set size
			Accused: uint64(ap.ApologizerIndex), //nolint:gosec // G115: keyper index is bounded by the keyper set size
			Eval:    eval,
		})
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-set-index", keyperSetIndex).
				Int64("retry-counter", retryCounter).
				Int64("apologizer", ap.ApologizerIndex).
				Int64("accuser", ap.AccuserIndex).
				Msg("ignoring apology on replay")
		}
	}
	return nil
}

func (m *Manager) decryptPolyEval(encrypted []byte) (*big.Int, error) {
	plaintext, err := m.cfg.ECIESPrivateKey.Decrypt(encrypted, []byte(""), []byte(""))
	if err != nil {
		return nil, errors.Wrap(err, "decrypt poly eval")
	}
	return new(big.Int).SetBytes(plaintext), nil
}

func (m *Manager) encryptPolyEvalFor(
	ctx context.Context,
	queries *corekeyperdb.Queries,
	receiverAddr common.Address,
	eval *big.Int,
) ([]byte, error) {
	row, err := queries.GetECIESKey(ctx, shdb.EncodeAddress(receiverAddr))
	if err != nil {
		return nil, errors.Wrapf(err, "look up ECIES key for %s", receiverAddr.Hex())
	}
	pubKey, err := shdb.DecodeEciesPublicKey(row.EciesPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "decode ECIES public key")
	}
	ciphertext, err := ecies.Encrypt(rand.Reader, pubKey, eval.Bytes(), []byte(""), []byte(""))
	if err != nil {
		return nil, errors.Wrap(err, "encrypt poly eval")
	}
	return ciphertext, nil
}

// eciesPrivateKeyFromECDSA constructs an `*ecies.PrivateKey` from a standard
// `*ecdsa.PrivateKey`. Callers without ready access to the ecies form can use
// this helper when building Config.
func eciesPrivateKeyFromECDSA(priv *ecdsa.PrivateKey) *ecies.PrivateKey {
	return ecies.ImportECDSA(priv)
}
