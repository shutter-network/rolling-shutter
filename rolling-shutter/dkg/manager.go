package dkg

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/shutter-network/shutter/shlib/puredkg"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/keypermetrics"
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

// HandleBlock is called once per new block by the host keyper. It asks
// `activeDKGs` which eons this keyper must still consider at this block, then
// dispatches each one to `processDKG`. Errors from individual eons are logged
// but do not abort the loop — a per-eon failure must not stop the others.
func (m *Manager) HandleBlock(ctx context.Context, blockNumber uint64) error {
	if blockNumber == 0 {
		return nil
	}
	m.publishMetrics(ctx, blockNumber)
	eons, summary, err := m.activeDKGs(ctx, blockNumber)
	if err != nil {
		return errors.Wrap(err, "list active DKGs")
	}
	log.Debug().
		Uint64("block-number", blockNumber).
		Int("active", len(eons)).
		Int("filtered-not-member", summary.notMember).
		Int("filtered-succeeded", summary.succeeded).
		Int("filtered-superseded", summary.superseded).
		Int("filtered-retries-exhausted", summary.retriesExhausted).
		Msg("DKG manager: active DKGs")
	for _, eon := range eons {
		if err := m.processDKG(ctx, eon, blockNumber); err != nil {
			log.Error().Err(err).
				Int64("keyper-set-index", eon.KeyperConfigIndex).
				Uint64("block-number", blockNumber).
				Msg("DKG manager: per-eon handler failed")
		}
	}
	return nil
}

// publishMetrics updates the Prometheus gauges the DKG manager owns for this
// block
func (m *Manager) publishMetrics(ctx context.Context, blockNumber uint64) {
	keypermetrics.MetricsKeyperCurrentBlockL1.Set(float64(blockNumber))

	keyperSets, err := obskeyper.New(m.cfg.DBPool).GetKeyperSets(ctx)
	if err != nil {
		log.Error().Err(err).Msg("DKG manager: failed to fetch keyper sets for is_keyper metric")
		return
	}
	for _, keyperSet := range keyperSets {
		isMember := 0.0
		if keyperSet.Contains(m.cfg.OwnAddress) {
			isMember = 1
		}
		keypermetrics.MetricsKeyperIsKeyper.
			WithLabelValues(strconv.FormatInt(keyperSet.KeyperConfigIndex, 10)).
			Set(isMember)
	}
}

// errNotMember is returned by `resolveMembership` when the manager's address
// is not present in the keyper set. Callers translate this sentinel into a
// silent no-op (matching the pre-refactor behavior of returning nil from
// `processDKG`).
var errNotMember = errors.New("not a member of keyper set")

// phaseParams bundles the eon fields required for DKG phase math. It is
// populated by `Manager.resolvePhaseParams` from the corresponding `eons`
// columns.
type phaseParams struct {
	activationBlock uint64
	phaseLength     uint64
	leadLength      uint64
	maxRetries      uint64
}

// keyperMembership bundles the keyper-set fields required to interact with
// `puredkg` on behalf of this keyper. It is populated by
// `Manager.resolveMembership` from a fetched `KeyperSet`.
type keyperMembership struct {
	keypers   []common.Address
	ownIndex  uint64
	threshold uint64
}

// resolvePhaseParams reads the DKG phase timing fields off an `eons` row and
// converts them to unsigned block-arithmetic types. Rows populated by
// `processNewKeyperSet` carry the values read from the keyper-set-specific
// DKG contract; rows with NULL `phase_length` or `lead_length` are a fatal
// configuration error — the module owns no chain client and there is no
// fallback. `max_retries` is a NOT NULL column so is always present; see the
// `MAX_RETRIES` glossary entry for its semantics.
func (m *Manager) resolvePhaseParams(eon corekeyperdb.Eon) (phaseParams, error) {
	if !eon.PhaseLength.Valid || !eon.LeadLength.Valid {
		return phaseParams{}, errors.Errorf(
			"eons row %d missing DKG phase params (phase_length and/or lead_length is NULL)",
			eon.KeyperConfigIndex,
		)
	}
	activationBlock, err := medley.Int64ToUint64Safe(eon.ActivationBlockNumber)
	if err != nil {
		return phaseParams{}, errors.Wrap(err, "convert activation block")
	}
	//nolint:gosec // G115: phase length, lead length, and max retries come from the on-chain contract and are non-negative
	return phaseParams{
		activationBlock: activationBlock,
		phaseLength:     uint64(eon.PhaseLength.Int64),
		leadLength:      uint64(eon.LeadLength.Int64),
		maxRetries:      uint64(eon.MaxRetries),
	}, nil
}

// resolveMembership derives this keyper's position in `keyperSet`. If the
// manager's address is not a member it returns `errNotMember`; callers turn
// that into a silent no-op. Any other error (address decoding, threshold
// conversion) is surfaced.
func (m *Manager) resolveMembership(keyperSet obskeyper.KeyperSet) (keyperMembership, error) {
	ownIndex, err := keyperSet.GetIndex(m.cfg.OwnAddress)
	if err != nil {
		return keyperMembership{}, errNotMember
	}
	keypers, err := shdb.DecodeAddresses(keyperSet.Keypers)
	if err != nil {
		return keyperMembership{}, errors.Wrap(err, "decode keyper addresses")
	}
	threshold, err := medley.Int32ToUint64Safe(keyperSet.Threshold)
	if err != nil {
		return keyperMembership{}, errors.Wrap(err, "convert keyper set threshold")
	}
	return keyperMembership{
		keypers:   keypers,
		ownIndex:  ownIndex,
		threshold: threshold,
	}, nil
}

// activeDKGsSummary counts how many eons were filtered out by each reason.
// Populated by `activeDKGs` and logged by `HandleBlock` as a single per-block
// summary line — there is no per-skip log line.
type activeDKGsSummary struct {
	notMember        int
	succeeded        int
	superseded       int
	retriesExhausted int
}

// activeDKGs returns the eons this keyper must still consider at `blockNumber`.
// It applies the "over or not mine" filters in one place so per-eon processing
// can focus on phase math. Filters, in the order applied — membership first
// because it is the most selective, then prior success, then supersession,
// then the retry ceiling:
//
//  1. Local keyper is a member of the eon's Keyper Set.
//  2. No `dkg_result` success row exists for the Keyper Set Index.
//  3. The Keyper Set is not superseded at this block (i.e. no later Keyper
//     Set has already activated). See CONTEXT.md#superseded-keyper-set.
//  4. The current retry counter is below `MAX_RETRIES`.
//
// The supersession predicate reuses the observer's existing "latest keyper
// set with activation block <= N" query (`GetKeyperSet`). `pgx.ErrNoRows` from
// that query means "no Keyper Set is live yet" and no set is considered
// superseded — future scheduled sets remain active.
//
// Eons whose configuration is too incomplete to evaluate a filter (e.g. NULL
// `phase_length`) are passed through: `processDKG` will surface the config
// error, HandleBlock will log it. A single query joining eons + keyper sets +
// dkg_result would be more efficient, but keyper sets live in the observer
// schema (separate connection pool) so cross-schema joins are not possible.
//
// The returned summary counts, one per filter, are populated so HandleBlock
// can emit a single per-block Debug summary line without any per-skip log
// entries. On error the returned summary is zero-valued.
func (m *Manager) activeDKGs(ctx context.Context, blockNumber uint64) ([]corekeyperdb.Eon, activeDKGsSummary, error) {
	queries := corekeyperdb.New(m.cfg.DBPool)
	obsQueries := obskeyper.New(m.cfg.DBPool)
	allEons, err := queries.GetAllEons(ctx)
	if err != nil {
		return nil, activeDKGsSummary{}, errors.Wrap(err, "list eons")
	}

	// Resolve the "latest keyper set at this block" once. Its index is the
	// supersession threshold: any Keyper Set with a strictly smaller index is
	// superseded. `pgx.ErrNoRows` means no set is live yet, so nothing is
	// superseded (future scheduled sets stay active).
	var latestActiveIndex int64
	haveLatestActive := false
	//nolint:gosec // G115: block numbers fit in int64 in practice.
	latestActive, err := obsQueries.GetKeyperSet(ctx, int64(blockNumber))
	switch {
	case err == nil:
		latestActiveIndex = latestActive.KeyperConfigIndex
		haveLatestActive = true
	case errors.Is(err, pgx.ErrNoRows):
		// No Keyper Set is live yet — no supersession possible.
	default:
		return nil, activeDKGsSummary{}, errors.Wrap(err, "fetch latest active keyper set")
	}

	summary := activeDKGsSummary{}
	active := make([]corekeyperdb.Eon, 0, len(allEons))
	for _, eon := range allEons {
		keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, eon.KeyperConfigIndex)
		if err != nil {
			return nil, activeDKGsSummary{}, errors.Wrapf(err, "fetch keyper set %d", eon.KeyperConfigIndex)
		}
		if _, err := m.resolveMembership(keyperSet); err != nil {
			if errors.Is(err, errNotMember) {
				summary.notMember++
				continue
			}
			return nil, activeDKGsSummary{}, err
		}
		alreadySucceeded, err := queries.ExistsDKGResultSuccess(ctx, eon.KeyperConfigIndex)
		if err != nil {
			return nil, activeDKGsSummary{}, errors.Wrap(err, "check existing dkg_result")
		}
		if alreadySucceeded {
			summary.succeeded++
			continue
		}
		if haveLatestActive && eon.KeyperConfigIndex < latestActiveIndex {
			summary.superseded++
			continue
		}
		// Eons missing phase params are passed through so that processDKG
		// surfaces the config error; the retry-ceiling check simply doesn't
		// apply in that case.
		if params, err := m.resolvePhaseParams(eon); err == nil {
			retry := CurrentRetryCounter(params.activationBlock, params.leadLength, params.phaseLength, blockNumber)
			if retry >= params.maxRetries {
				summary.retriesExhausted++
				continue
			}
		}
		active = append(active, eon)
	}
	return active, summary, nil
}

// processDKG runs the per-block phase math and dispatches at most one action
// for a single eon. It assumes `activeDKGs` has already filtered out eons
// that this keyper is not a member of, that have already succeeded, or whose
// retry counter has reached MAX_RETRIES.
//
// Pool queries (no transaction) fire first: phase params and the keyper-set
// lookup needed for `ownIndex`. Once dispatch is required, two narrow
// transactions are opened in sequence: a read-only one for `buildPureDKG`,
// then a separate write one for the maybe-function. Splitting the transactions
// keeps the read window short and lets the write transaction commit
// independently. A new chain event arriving between the two transactions is
// acceptable — the maybe-function will see the stale snapshot for one block
// and pick up the new state on the next dispatch.
//
// Returns nil for "nothing to do" (no active phase at this block, no initial
// state for non-Dealing phases). Returns an error for missing per-eon
// configuration (NULL `dkg_contract` or NULL `phase_length`/`lead_length`).
// The caller logs but does not abort on error.
func (m *Manager) processDKG(ctx context.Context, eon corekeyperdb.Eon, blockNumber uint64) error {
	params, err := m.resolvePhaseParams(eon)
	if err != nil {
		return err
	}

	obsQueries := obskeyper.New(m.cfg.DBPool)
	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, eon.KeyperConfigIndex)
	if err != nil {
		return errors.Wrapf(err, "fetch keyper set %d", eon.KeyperConfigIndex)
	}
	member, err := m.resolveMembership(keyperSet)
	if err != nil {
		if errors.Is(err, errNotMember) {
			// Belt-and-braces: activeDKGs already filters non-members, but if
			// a caller invokes processDKG directly (tests, future diagnostics)
			// we exit silently rather than surface a misleading error.
			return nil
		}
		return err
	}

	retry := CurrentRetryCounter(params.activationBlock, params.leadLength, params.phaseLength, blockNumber)
	retryInt64 := int64(retry) //nolint:gosec // G115: retry counter is bounded by the on-chain contract
	blockPhase := PhaseAt(params.activationBlock, params.leadLength, params.phaseLength, params.maxRetries, retry, blockNumber)
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
		p, err := m.buildPureDKG(ctx, tx, eon.KeyperConfigIndex, retryInt64, blockPhase, member.keypers, member.ownIndex, member.threshold)
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
			return m.maybeDeal(ctx, tx, dkgAddr, eon.KeyperConfigIndex, retryInt64, pure, member.keypers, member.ownIndex)
		case PhaseAccusing:
			return m.maybeAccuse(ctx, tx, dkgAddr, eon.KeyperConfigIndex, retryInt64, pure, member.keypers, member.ownIndex)
		case PhaseApologizing:
			return m.maybeApologize(ctx, tx, dkgAddr, eon.KeyperConfigIndex, retryInt64, pure, member.keypers, member.ownIndex)
		case PhaseFinalizing:
			return m.maybeFinalize(ctx, tx, dkgAddr, eon.KeyperConfigIndex, retryInt64, pure, member.ownIndex)
		case PhaseNone:
			return nil
		default:
			return nil
		}
	})
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
func (m *Manager) HandleDKGSuccess(ctx context.Context, tx pgx.Tx, keyperSetIndex, retryCounter int64) error {
	queries := corekeyperdb.New(tx)
	exists, err := queries.ExistsDKGResultSuccess(ctx, keyperSetIndex)
	if err != nil {
		return errors.Wrap(err, "check existing dkg_result")
	}
	if exists {
		return nil
	}

	obsQueries := obskeyper.New(tx)
	keyperSet, err := obsQueries.GetKeyperSetByKeyperConfigIndex(ctx, keyperSetIndex)
	if err != nil {
		return errors.Wrapf(err, "fetch keyper set %d", keyperSetIndex)
	}

	var pureBytes []byte
	var resultNumKeypers, resultThreshold uint64
	var accusationCount, apologyCount int
	var hasResult bool
	var localErr sql.NullString

	member, memberErr := m.resolveMembership(keyperSet)
	if memberErr != nil && !errors.Is(memberErr, errNotMember) {
		return memberErr
	}
	if memberErr == nil {
		pure, err := m.buildPureDKG(ctx, tx, keyperSetIndex, retryCounter, PhaseFinalizing, member.keypers, member.ownIndex, member.threshold)
		if err != nil {
			return errors.Wrap(err, "rebuild puredkg for success")
		}
		if pure == nil {
			localErr = sql.NullString{String: "local: buildPureDKG returned nil", Valid: true}
		} else {
			pure.Finalize()
			result, err := pure.ComputeResult()
			if err != nil {
				localErr = sql.NullString{String: fmt.Sprintf("local: compute result failed: %s", err), Valid: true}
				log.Warn().Err(err).
					Int64("keyper-set-index", keyperSetIndex).
					Int64("retry-counter", retryCounter).
					Msg("cannot compute DKG result on success event; storing nil")
			} else {
				hasResult = true
				resultNumKeypers = result.NumKeypers
				resultThreshold = result.Threshold
				accusationCount = len(pure.Accusations)
				apologyCount = len(pure.Apologies)
				pureBytes, err = shdb.EncodePureDKGResult(&result)
				if err != nil {
					return errors.Wrap(err, "encode pure DKG result")
				}
			}
		}
	}

	event := log.Info().
		Int64("keyper-set-index", keyperSetIndex).
		Int64("retry-counter", retryCounter)
	if hasResult {
		event = event.
			Uint64("keyper-count", resultNumKeypers).
			Uint64("threshold", resultThreshold).
			Int("accusation-count", accusationCount).
			Int("apology-count", apologyCount)
	}
	event.Msg("DKG succeeded")
	return queries.InsertDKGResult(ctx, corekeyperdb.InsertDKGResultParams{
		Eon:        keyperSetIndex,
		Success:    true,
		Error:      localErr,
		PureResult: pureBytes,
	})
}
