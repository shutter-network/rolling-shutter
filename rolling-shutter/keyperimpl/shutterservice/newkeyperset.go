package shutterservice

import (
	"context"
	"database/sql"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/contracts/v2/bindings/dkgcontract"
	keypersetBindings "github.com/shutter-network/contracts/v2/bindings/keyperset"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdb "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func (kpr *Keyper) processNewKeyperSet(ctx context.Context, ev *syncevent.KeyperSet) error {
	ownAddress := kpr.config.GetAddress()
	isMember := false
	for _, m := range ev.Members {
		if m.Cmp(ownAddress) == 0 {
			isMember = true
			break
		}
	}

	log.Info().
		Uint64("activation-block", ev.ActivationBlock).
		Uint64("eon", ev.Eon).
		Int("num-members", len(ev.Members)).
		Uint64("threshold", ev.Threshold).
		Bool("is-member", isMember).
		Msg("new keyper set added")

	keyperSetIndex, err := medley.Uint64ToInt64Safe(ev.Eon)
	if err != nil {
		return errors.Wrap(err, ErrParseKeyperSet.Error())
	}

	if err := kpr.dbpool.BeginFunc(ctx, func(tx pgx.Tx) error {
		obskeyperdb := obskeyper.New(tx)
		coredb := corekeyperdb.New(tx)

		activationBlockNumber, err := medley.Uint64ToInt64Safe(ev.ActivationBlock)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}
		threshold, err := medley.Uint64ToInt64Safe(ev.Threshold)
		if err != nil {
			return errors.Wrap(err, ErrParseKeyperSet.Error())
		}

		if err := obskeyperdb.InsertKeyperSet(ctx, obskeyper.InsertKeyperSetParams{
			KeyperConfigIndex:     keyperSetIndex,
			ActivationBlockNumber: activationBlockNumber,
			Keypers:               shdb.EncodeAddresses(ev.Members),
			Threshold:             int32(threshold),
		}); err != nil {
			return err
		}

		if isMember {
			// Eagerly insert an eons row so the DKG participation loop has
			// somewhere to anchor when the activation block approaches.
			// Existing rows are tolerated because the chainsync initial poll
			// can re-deliver KeyperSetAdded events that were already processed.
			if _, err := coredb.GetEon(ctx, keyperSetIndex); err == nil {
				return nil
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return errors.Wrap(err, "check existing eon row")
			}
			dkgContract, phaseLength, leadLength := kpr.fetchDKGParamsForKeyperSet(ctx, ev.Contract)
			if err := coredb.InsertEon(ctx, corekeyperdb.InsertEonParams{
				Eon:                   keyperSetIndex,
				ActivationBlockNumber: activationBlockNumber,
				KeyperConfigIndex:     keyperSetIndex,
				DkgContract:           dkgContract,
				PhaseLength:           phaseLength,
				LeadLength:            leadLength,
			}); err != nil {
				return errors.Wrap(err, "insert eon row for new keyper set")
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// ECIES key registration runs at the same architectural level as
	// HandleBlock — once per discovered keyper set, after the keyper set
	// row is committed (MaybeRegisterECIESKey reads it). It is idempotent
	// and a no-op for non-members.
	return kpr.dkgMgr.MaybeRegisterECIESKey(ctx, keyperSetIndex)
}

// fetchDKGParamsForKeyperSet asks the keyper set contract for its DKG contract
// address and then reads the immutable phase parameters from that DKG contract.
// Any failure (zero address, RPC error, missing methods on an old contract) is
// logged and surfaced as NULL columns; downstream callers fall back to the
// config-supplied DKG contract address. This is a best-effort enrichment, not
// a precondition for joining the keyper set.
func (kpr *Keyper) fetchDKGParamsForKeyperSet(
	ctx context.Context,
	keyperSetAddr common.Address,
) (sql.NullString, sql.NullInt64, sql.NullInt64) {
	var (
		nullStr sql.NullString
		nullInt sql.NullInt64
	)
	if (keyperSetAddr == common.Address{}) {
		log.Warn().Msg("keyper set event missing contract address; storing NULL phase params")
		return nullStr, nullInt, nullInt
	}
	ks, err := keypersetBindings.NewKeyperset(keyperSetAddr, kpr.chainSyncClient.Client)
	if err != nil {
		log.Warn().Err(err).Str("keyper-set", keyperSetAddr.Hex()).
			Msg("bind keyper set contract for DKG lookup; storing NULL phase params")
		return nullStr, nullInt, nullInt
	}
	callOpts := &bind.CallOpts{Context: ctx}
	dkgAddr, err := ks.GetDKGContract(callOpts)
	if err != nil {
		log.Warn().Err(err).Str("keyper-set", keyperSetAddr.Hex()).
			Msg("call getDKGContract; storing NULL phase params")
		return nullStr, nullInt, nullInt
	}
	if (dkgAddr == common.Address{}) {
		log.Warn().Str("keyper-set", keyperSetAddr.Hex()).
			Msg("keyper set has no DKG contract configured; storing NULL phase params")
		return nullStr, nullInt, nullInt
	}
	dkg, err := dkgcontract.NewDkgcontract(dkgAddr, kpr.chainSyncClient.Client)
	if err != nil {
		log.Warn().Err(err).Str("dkg-contract", dkgAddr.Hex()).
			Msg("bind DKG contract; storing NULL phase params")
		return nullStr, nullInt, nullInt
	}
	phaseLength, err := dkg.PHASELENGTH(callOpts)
	if err != nil {
		log.Warn().Err(err).Str("dkg-contract", dkgAddr.Hex()).
			Msg("read PHASE_LENGTH; storing NULL phase params")
		return nullStr, nullInt, nullInt
	}
	leadLength, err := dkg.DKGLEADLENGTH(callOpts)
	if err != nil {
		log.Warn().Err(err).Str("dkg-contract", dkgAddr.Hex()).
			Msg("read DKG_LEAD_LENGTH; storing NULL phase params")
		return nullStr, nullInt, nullInt
	}
	log.Info().
		Str("keyper-set", keyperSetAddr.Hex()).
		Str("dkg-contract", dkgAddr.Hex()).
		Uint64("phase-length", phaseLength).
		Uint64("lead-length", leadLength).
		Msg("resolved per-keyper-set DKG contract params")
	//nolint:gosec // G115: phase and lead lengths come from the on-chain contract and fit well within int64
	return sql.NullString{String: dkgAddr.Hex(), Valid: true},
		sql.NullInt64{Int64: int64(phaseLength), Valid: true},
		sql.NullInt64{Int64: int64(leadLength), Valid: true}
}
