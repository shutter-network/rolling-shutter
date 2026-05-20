package gnosis

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"
	"github.com/shutter-network/shutter/shlib/shcrypto"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

type dkgInstanceKey struct {
	keyperConfigIndex int64
	retryCounter      int64
}

// dkgInstance keeps the in-memory `puredkg` state for a single DKG attempt,
// together with the keyper-set context needed to interpret messages. The DB
// is the source of truth; this struct is a derived cache rebuilt on startup
// and updated as live messages arrive.
type dkgInstance struct {
	pure      *puredkg.PureDKG
	keypers   []common.Address
	ownIndex  uint64
	threshold uint64
}

type dkgManager struct {
	mu        sync.Mutex
	instances map[dkgInstanceKey]*dkgInstance
}

func newDKGManager() *dkgManager {
	return &dkgManager{
		instances: make(map[dkgInstanceKey]*dkgInstance),
	}
}

func (m *dkgManager) get(k dkgInstanceKey) *dkgInstance {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.instances[k]
}

func (m *dkgManager) put(k dkgInstanceKey, inst *dkgInstance) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instances[k] = inst
}

// loadOrBuildInstance returns the in-memory DKG state for the given instance
// key, building it from stored messages if it isn't cached yet. Returns nil
// when the keyper is not a member of the corresponding keyper set.
func (kpr *Keyper) loadOrBuildInstance(
	ctx context.Context,
	tx pgx.Tx,
	keyperConfigIndex, retryCounter int64,
) (*dkgInstance, error) {
	key := dkgInstanceKey{keyperConfigIndex: keyperConfigIndex, retryCounter: retryCounter}
	if existing := kpr.dkgManager.get(key); existing != nil {
		return existing, nil
	}

	obsQueries := obskeyper.New(tx)
	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, keyperConfigIndex)
	if err != nil {
		return nil, errors.Wrapf(err, "fetch keyper set %d", keyperConfigIndex)
	}
	ownAddr := kpr.config.GetAddress()
	ownIndex, err := keyperSet.GetIndex(ownAddr)
	if err != nil {
		return nil, nil
	}
	keypers, err := shdb.DecodeAddresses(keyperSet.Keypers)
	if err != nil {
		return nil, errors.Wrap(err, "decode keyper addresses")
	}

	pure := puredkg.NewPureDKG(
		uint64(keyperConfigIndex),
		uint64(len(keypers)),
		uint64(keyperSet.Threshold),
		ownIndex,
	)
	inst := &dkgInstance{
		pure:      &pure,
		keypers:   keypers,
		ownIndex:  ownIndex,
		threshold: uint64(keyperSet.Threshold),
	}
	if err := kpr.replayStoredMessages(ctx, tx, inst, keyperConfigIndex, retryCounter); err != nil {
		return nil, err
	}
	kpr.dkgManager.put(key, inst)
	return inst, nil
}

// replayStoredMessages feeds every stored DKG message for the given instance
// back into the in-memory puredkg, using puredkg's public `Handle*Msg` API.
// All four `Handle*` methods accept input while `pure.Phase` is at or below
// their target phase, so leaving the in-memory state at `Phase = Off` until
// the live phase boundary fires is correct here.
func (kpr *Keyper) replayStoredMessages(
	ctx context.Context,
	tx pgx.Tx,
	inst *dkgInstance,
	keyperConfigIndex, retryCounter int64,
) error {
	queries := corekeyperdb.New(tx)
	eonForMsg := uint64(keyperConfigIndex)

	commitments, err := queries.GetDKGPolyCommitments(ctx, corekeyperdb.GetDKGPolyCommitmentsParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	if err != nil {
		return errors.Wrap(err, "load stored poly commitments")
	}
	for _, c := range commitments {
		gammas := &shcrypto.Gammas{}
		if err := gammas.Unmarshal(c.Commitment); err != nil {
			return errors.Wrapf(err, "decode commitment from sender %d", c.KeyperIndex)
		}
		err := inst.pure.HandlePolyCommitmentMsg(puredkg.PolyCommitmentMsg{
			Eon:    eonForMsg,
			Sender: uint64(c.KeyperIndex),
			Gammas: gammas,
		})
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-config-index", keyperConfigIndex).
				Int64("retry-counter", retryCounter).
				Int64("sender", c.KeyperIndex).
				Msg("ignoring stored commitment on replay")
		}
	}

	polyEvals, err := queries.GetDKGPolyEvals(ctx, corekeyperdb.GetDKGPolyEvalsParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	if err != nil {
		return errors.Wrap(err, "load stored poly evals")
	}
	for _, ev := range polyEvals {
		if uint64(ev.ReceiverIndex) != inst.ownIndex {
			continue
		}
		eval, err := kpr.decryptPolyEval(ev.EncryptedEval)
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-config-index", keyperConfigIndex).
				Int64("retry-counter", retryCounter).
				Int64("sender", ev.SenderIndex).
				Msg("ignoring undecryptable poly eval on replay")
			continue
		}
		err = inst.pure.HandlePolyEvalMsg(puredkg.PolyEvalMsg{
			Eon:      eonForMsg,
			Sender:   uint64(ev.SenderIndex),
			Receiver: inst.ownIndex,
			Eval:     eval,
		})
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-config-index", keyperConfigIndex).
				Int64("retry-counter", retryCounter).
				Int64("sender", ev.SenderIndex).
				Msg("ignoring poly eval on replay")
		}
	}

	accusations, err := queries.GetDKGAccusations(ctx, corekeyperdb.GetDKGAccusationsParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	if err != nil {
		return errors.Wrap(err, "load stored accusations")
	}
	for _, a := range accusations {
		err := inst.pure.HandleAccusationMsg(puredkg.AccusationMsg{
			Eon:     eonForMsg,
			Accuser: uint64(a.AccuserIndex),
			Accused: uint64(a.AccusedIndex),
		})
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-config-index", keyperConfigIndex).
				Int64("retry-counter", retryCounter).
				Int64("accuser", a.AccuserIndex).
				Int64("accused", a.AccusedIndex).
				Msg("ignoring accusation on replay")
		}
	}

	apologies, err := queries.GetDKGApologies(ctx, corekeyperdb.GetDKGApologiesParams{
		KeyperConfigIndex: keyperConfigIndex,
		RetryCounter:      retryCounter,
	})
	if err != nil {
		return errors.Wrap(err, "load stored apologies")
	}
	for _, ap := range apologies {
		eval := new(big.Int).SetBytes(ap.PolyEval)
		err := inst.pure.HandleApologyMsg(puredkg.ApologyMsg{
			Eon:     eonForMsg,
			Accuser: uint64(ap.AccuserIndex),
			Accused: uint64(ap.ApologizerIndex),
			Eval:    eval,
		})
		if err != nil {
			log.Debug().Err(err).
				Int64("keyper-config-index", keyperConfigIndex).
				Int64("retry-counter", retryCounter).
				Int64("apologizer", ap.ApologizerIndex).
				Int64("accuser", ap.AccuserIndex).
				Msg("ignoring apology on replay")
		}
	}

	return nil
}

func (kpr *Keyper) decryptPolyEval(encrypted []byte) (*big.Int, error) {
	priv := kpr.config.Shuttermint.EncryptionKey.Key
	eciesPriv := ecies.ImportECDSA(priv)
	plaintext, err := eciesPriv.Decrypt(encrypted, []byte(""), []byte(""))
	if err != nil {
		return nil, errors.Wrap(err, "decrypt poly eval")
	}
	return new(big.Int).SetBytes(plaintext), nil
}

func (kpr *Keyper) encryptPolyEvalFor(
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
