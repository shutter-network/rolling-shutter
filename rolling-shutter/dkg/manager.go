package dkg

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
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
	// DKGContractAddr is the fallback DKG contract address for eons whose
	// `eons.dkg_contract` column is NULL (older databases or rows where the
	// per-keyper-set DKG contract lookup failed at insert time).
	DKGContractAddr common.Address
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
	dkgContractAddr common.Address,
	eciesRegistryAddr common.Address,
) Config {
	return Config{
		DBPool:            dbPool,
		OwnAddress:        ownAddress,
		ECIESPrivateKey:   eciesPrivateKeyFromECDSA(eciesECDSAPrivateKey),
		DKGContractAddr:   dkgContractAddr,
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

// HandleBlock is called once per new block by the host keyper. For every
// eon row in the database it dispatches all action handlers unconditionally:
// each handler's idempotency check (presence of an own message row in the
// relevant DKG table, or presence of a `dkg_result` row, or the ECIES key
// cache) makes repeated invocations safe and cheap.
//
// Errors from individual eons are logged but do not abort the loop — a
// per-eon failure must not stop the others.
func (m *Manager) HandleBlock(ctx context.Context, blockNumber uint64) error {
	if blockNumber == 0 {
		return nil
	}
	eons, err := corekeyperdb.New(m.cfg.DBPool).GetAllEons(ctx)
	if err != nil {
		return errors.Wrap(err, "list eons")
	}
	for _, eon := range eons {
		activationBlock, err := medley.Int64ToUint64Safe(eon.ActivationBlockNumber)
		if err != nil {
			log.Error().Err(err).
				Int64("keyper-config-index", eon.KeyperConfigIndex).
				Msg("DKG manager: convert activation block")
			continue
		}
		phaseLength, leadLength := m.phaseParamsForEon(eon)
		if phaseLength == 0 {
			continue
		}
		retry := CurrentRetryCounter(activationBlock, leadLength, phaseLength, blockNumber)
		dkgAddr := m.dkgContractAddrForEon(eon)

		// ECIES registration is independent of phase and runs once per
		// (eon, own keyper) pair until the registry confirms.
		if err := m.maybeRegisterECIESKey(ctx, eon); err != nil {
			log.Error().Err(err).
				Int64("keyper-config-index", eon.KeyperConfigIndex).
				Msg("DKG manager: ECIES key registration")
		}

		// The four DKG actions all run on every block; each is idempotent
		// via DB-state checks. Operate in dependency order: dealing, then
		// the message exchange, then finalisation.
		retryInt64 := int64(retry)
		if err := m.maybeDeal(ctx, dkgAddr, eon.KeyperConfigIndex, retryInt64); err != nil {
			log.Error().Err(err).
				Int64("keyper-config-index", eon.KeyperConfigIndex).
				Uint64("retry-counter", retry).
				Msg("DKG manager: dealing action failed")
		}
		if err := m.startAccusing(ctx, dkgAddr, eon.KeyperConfigIndex, retryInt64); err != nil {
			log.Error().Err(err).
				Int64("keyper-config-index", eon.KeyperConfigIndex).
				Uint64("retry-counter", retry).
				Msg("DKG manager: accusing action failed")
		}
		if err := m.startApologizing(ctx, dkgAddr, eon.KeyperConfigIndex, retryInt64); err != nil {
			log.Error().Err(err).
				Int64("keyper-config-index", eon.KeyperConfigIndex).
				Uint64("retry-counter", retry).
				Msg("DKG manager: apologizing action failed")
		}
		if err := m.startFinalizing(ctx, dkgAddr, eon.KeyperConfigIndex, retryInt64); err != nil {
			log.Error().Err(err).
				Int64("keyper-config-index", eon.KeyperConfigIndex).
				Uint64("retry-counter", retry).
				Msg("DKG manager: finalizing action failed")
		}
	}
	return nil
}

// phaseParamsForEon returns the DKG phase length and lead length for the
// given eon. Rows populated by `processNewKeyperSet` carry the values read
// from the keyper-set-specific DKG contract; rows from older databases (or
// rows where the lookup failed) carry NULLs and are skipped — there is no
// global fallback in the module because the module owns no chain client to
// fetch one.
func (m *Manager) phaseParamsForEon(eon corekeyperdb.Eon) (phaseLength, leadLength uint64) {
	if eon.PhaseLength.Valid && eon.LeadLength.Valid {
		return uint64(eon.PhaseLength.Int64), uint64(eon.LeadLength.Int64)
	}
	log.Warn().
		Int64("keyper-config-index", eon.KeyperConfigIndex).
		Msg("eons row missing DKG phase params; skipping DKG actions for this eon")
	return 0, 0
}

// dkgContractAddrForEon returns the on-chain DKG contract address responsible
// for the given eon. Rows where the per-keyper-set lookup succeeded carry the
// concrete address; rows with NULL fall back to the manager's configured
// `DKGContractAddr`.
func (m *Manager) dkgContractAddrForEon(eon corekeyperdb.Eon) common.Address {
	if eon.DkgContract.Valid {
		return common.HexToAddress(eon.DkgContract.String)
	}
	return m.cfg.DKGContractAddr
}
