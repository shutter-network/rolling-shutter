package dkg

import (
	"context"
	"crypto/ecdsa"
	"database/sql"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

// Action names recorded in the `dkg_sent_actions` table by each reactor on
// the success path. The table is the uniform idempotency store across all
// four reactors.
const (
	ActionDealing     = "dealing"
	ActionAccusing    = "accusing"
	ActionApologizing = "apologizing"
	ActionFinalizing  = "finalizing"
)

// Config carries the dependencies needed to run a DKG Manager. The manager
// reads chain state exclusively through the database (populated by the host
// keyper's chain syncers) and writes both message rows and outbox tx intents
// to the same database. It never holds a chain client.
type Config struct {
	// DBPool is the database connection pool. Required.
	DBPool *pgxpool.Pool
	// OwnAddress identifies the keyper for DKG membership checks. Required.
	OwnAddress common.Address
	// ECIESPrivateKey is the keyper's ECIES private key used to decrypt
	// PolyEval messages addressed to us.
	ECIESPrivateKey *ecies.PrivateKey
	// ECIESRegistryAddr is the ECIES key registry contract address. Used as
	// the destination for the `registerKey` outbox entry on first
	// participation in a keyper set.
	ECIESRegistryAddr common.Address
}

// NewConfigFromECDSA is a small convenience wrapper that builds a Config
// from a standard ECDSA private key without the caller needing to import the
// `crypto/ecies` package itself.
func NewConfigFromECDSA(
	dbPool *pgxpool.Pool,
	ownAddress common.Address,
	eciesECDSAPrivateKey *ecdsa.PrivateKey,
	eciesRegistryAddr common.Address,
) Config {
	return Config{
		DBPool:            dbPool,
		OwnAddress:        ownAddress,
		ECIESPrivateKey:   eciesPrivateKeyFromECDSA(eciesECDSAPrivateKey),
		ECIESRegistryAddr: eciesRegistryAddr,
	}
}

// Manager owns no goroutines, no chain subscriptions, and no in-memory
// caches. HandleBlock is the only entry point; it is called by the host
// keyper on each new block.
type Manager struct {
	cfg Config
}

// New constructs a Manager with the given configuration.
func New(cfg Config) *Manager {
	return &Manager{cfg: cfg}
}

// HandleBlock is called once per new block by the host keyper. For every eon
// row in the database it computes the current DKG phase and dispatches to at
// most one action function. Errors from individual eons are logged but do
// not abort the loop — a per-eon failure must not stop the others.
func (m *Manager) HandleBlock(ctx context.Context, blockNumber uint64) error {
	if blockNumber == 0 {
		return nil
	}
	// Fetches all eons then filters per-eon (membership, dkg_result). A single
	// query joining eons + keyper sets + dkg_result would be more efficient,
	// but keyper sets live in the observer schema (separate connection pool),
	// so cross-schema joins are not possible here.
	queries := corekeyperdb.New(m.cfg.DBPool)
	eons, err := queries.GetAllEons(ctx)
	if err != nil {
		return errors.Wrap(err, "list eons")
	}
	for _, eon := range eons {
		if err := m.handleEon(ctx, eon, blockNumber); err != nil {
			log.Error().Err(err).
				Int64("keyper-config-index", eon.KeyperConfigIndex).
				Uint64("block-number", blockNumber).
				Msg("DKG manager: per-eon handler failed")
		}
	}
	return nil
}

// handleEon runs the per-eon dispatch logic. Pool queries (no transaction)
// fire first: phase params, membership, and `ExistsDKGResultSuccess`.
// Once dispatch is required, two narrow transactions are opened in
// sequence: a read-only one for `buildPureDKG`, then a separate write one
// for the maybe-function. Splitting the transactions keeps the read window
// short and lets the write transaction commit independently. A new chain
// event arriving between the two transactions is acceptable — the
// maybe-function will see the stale snapshot for one block and pick up the
// new state on the next dispatch.
//
// Returns nil for "nothing to do" (not a member, eon already succeeded,
// no active phase at this block, no initial state for non-Dealing phases).
// Returns an error for missing per-eon configuration (NULL `dkg_contract`
// or NULL `phase_length`/`lead_length`). The caller logs but does not
// abort on error.
func (m *Manager) handleEon(ctx context.Context, eon corekeyperdb.Eon, blockNumber uint64) error {
	activationBlock, err := medley.Int64ToUint64Safe(eon.ActivationBlockNumber)
	if err != nil {
		return errors.Wrap(err, "convert activation block")
	}
	phaseLength, leadLength, err := m.phaseParamsForEon(eon)
	if err != nil {
		return err
	}

	// Early exit: skip eons whose keyper set does not include this keyper.
	// Membership is the most selective filter — fire it before any other
	// DB work so we do not query dkg_result for unrelated sets.
	obsQueries := obskeyper.New(m.cfg.DBPool)
	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, eon.KeyperConfigIndex)
	if err != nil {
		return errors.Wrapf(err, "fetch keyper set %d", eon.KeyperConfigIndex)
	}
	ownIndex, err := keyperSet.GetIndex(m.cfg.OwnAddress)
	if err != nil {
		return nil
	}
	keypers, err := shdb.DecodeAddresses(keyperSet.Keypers)
	if err != nil {
		return errors.Wrap(err, "decode keyper addresses")
	}
	threshold, err := medley.Int32ToUint64Safe(keyperSet.Threshold)
	if err != nil {
		return errors.Wrap(err, "convert keyper set threshold")
	}

	// Early exit: a successful dkg_result row means either we already voted
	// or the chain has concluded the DKG. Nothing more to do for this eon.
	queries := corekeyperdb.New(m.cfg.DBPool)
	alreadySucceeded, err := queries.ExistsDKGResultSuccess(ctx, eon.KeyperConfigIndex)
	if err != nil {
		return errors.Wrap(err, "check existing dkg_result")
	}
	if alreadySucceeded {
		return nil
	}

	retry := CurrentRetryCounter(activationBlock, leadLength, phaseLength, blockNumber)
	retryInt64 := int64(retry)
	blockPhase := PhaseAt(activationBlock, leadLength, phaseLength, retry, blockNumber)
	if blockPhase == PhaseNone {
		return nil
	}
	dkgAddr, err := m.dkgContractAddrForEon(eon)
	if err != nil {
		return err
	}

	// Read transaction: rebuild the puredkg snapshot. Reads only; commit and
	// rollback are equivalent here, so we let BeginFunc commit on nil return.
	var pure *puredkg.PureDKG
	err = m.cfg.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		p, err := m.buildPureDKG(ctx, tx, eon.KeyperConfigIndex, retryInt64, blockPhase, keypers, ownIndex, threshold)
		if err != nil {
			return errors.Wrap(err, "build puredkg")
		}
		pure = p
		return nil
	})
	if err != nil {
		return err
	}
	if pure == nil {
		return nil
	}

	// Write transaction: dispatch to the maybe-function. Scope is narrow —
	// only the outbox row, idempotency rows, and `dkg_initial_states` /
	// `dkg_result` writes happen here.
	return m.cfg.DBPool.BeginFunc(ctx, func(tx pgx.Tx) error {
		switch blockPhase {
		case PhaseDealing:
			return m.maybeDeal(ctx, tx, dkgAddr, eon.KeyperConfigIndex, retryInt64, pure, keypers, ownIndex)
		case PhaseAccusing:
			return m.maybeAccuse(ctx, tx, dkgAddr, eon.KeyperConfigIndex, retryInt64, pure, ownIndex)
		case PhaseApologizing:
			return m.maybeApologize(ctx, tx, dkgAddr, eon.KeyperConfigIndex, retryInt64, pure, ownIndex)
		case PhaseFinalizing:
			return m.maybeFinalize(ctx, tx, dkgAddr, eon.KeyperConfigIndex, retryInt64, pure, ownIndex)
		default:
			return nil
		}
	})
}

// phaseParamsForEon returns the DKG phase length and lead length for the
// given eon. Rows populated by `processNewKeyperSet` carry the values read
// from the keyper-set-specific DKG contract; rows with NULL columns are a
// fatal configuration error — the module owns no chain client and there is
// no fallback to fetch them from.
func (m *Manager) phaseParamsForEon(eon corekeyperdb.Eon) (phaseLength, leadLength uint64, err error) {
	if !eon.PhaseLength.Valid || !eon.LeadLength.Valid {
		return 0, 0, errors.Errorf(
			"eons row %d missing DKG phase params (phase_length and/or lead_length is NULL)",
			eon.KeyperConfigIndex,
		)
	}
	return uint64(eon.PhaseLength.Int64), uint64(eon.LeadLength.Int64), nil
}

// dkgContractAddrForEon returns the on-chain DKG contract address responsible
// for the given eon. A NULL `dkg_contract` column is a fatal configuration
// error — there is no global fallback that could mask a misconfigured keyper
// set.
func (m *Manager) dkgContractAddrForEon(eon corekeyperdb.Eon) (common.Address, error) {
	if !eon.DkgContract.Valid {
		return common.Address{}, errors.Errorf(
			"eons row %d missing DKG contract address (dkg_contract is NULL)",
			eon.KeyperConfigIndex,
		)
	}
	return common.HexToAddress(eon.DkgContract.String), nil
}

// HandleDKGSuccess is called by the chain-event handler when a DKGSucceeded
// event arrives on chain. It is the single writer of `dkg_result` rows: the
// chain event is the source of truth for which retry actually won.
//
// If this keyper participated in `retryCounter` it rebuilds the puredkg state
// and stores the computed result in `pure_result`; otherwise `pure_result` is
// nil. Both outcomes produce a `dkg_result` row with `success=true`.
func (m *Manager) HandleDKGSuccess(ctx context.Context, tx pgx.Tx, keyperConfigIndex, retryCounter int64) error {
	queries := corekeyperdb.New(tx)
	exists, err := queries.ExistsDKGResultSuccess(ctx, keyperConfigIndex)
	if err != nil {
		return errors.Wrap(err, "check existing dkg_result")
	}
	if exists {
		return nil
	}

	obsQueries := obskeyper.New(tx)
	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, keyperConfigIndex)
	if err != nil {
		return errors.Wrapf(err, "fetch keyper set %d", keyperConfigIndex)
	}

	var pureBytes []byte
	ownIndex, memberErr := keyperSet.GetIndex(m.cfg.OwnAddress)
	if memberErr == nil {
		keypers, err := shdb.DecodeAddresses(keyperSet.Keypers)
		if err != nil {
			return errors.Wrap(err, "decode keyper addresses")
		}
		threshold, err := medley.Int32ToUint64Safe(keyperSet.Threshold)
		if err != nil {
			return errors.Wrap(err, "convert threshold")
		}
		pure, err := m.buildPureDKG(ctx, tx, keyperConfigIndex, retryCounter, PhaseFinalizing, keypers, ownIndex, threshold)
		if err != nil {
			return errors.Wrap(err, "rebuild puredkg for success")
		}
		if pure != nil {
			pure.Finalize()
			result, err := pure.ComputeResult()
			if err != nil {
				log.Warn().Err(err).
					Int64("keyper-config-index", keyperConfigIndex).
					Int64("retry-counter", retryCounter).
					Msg("cannot compute DKG result on success event; storing nil")
			} else {
				pureBytes, err = shdb.EncodePureDKGResult(&result)
				if err != nil {
					return errors.Wrap(err, "encode pure DKG result")
				}
			}
		}
	}

	log.Info().
		Int64("keyper-config-index", keyperConfigIndex).
		Int64("retry-counter", retryCounter).
		Msg("recording DKG success")
	return queries.InsertDKGResult(ctx, corekeyperdb.InsertDKGResultParams{
		Eon:        keyperConfigIndex,
		Success:    true,
		Error:      sql.NullString{},
		PureResult: pureBytes,
	})
}
