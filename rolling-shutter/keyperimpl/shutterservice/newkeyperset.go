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

	// Fetch DKG phase params before opening the DB transaction: the RPC calls
	// can be slow and we do not want to hold row locks across them. On event
	// replay this is a wasted round-trip because the existing-row check inside
	// the transaction will short-circuit, but replays are rare.
	var (
		dkgContract sql.NullString
		phaseLength sql.NullInt64
		leadLength  sql.NullInt64
		maxRetries  int64
	)
	if isMember {
		dkgContract, phaseLength, leadLength, maxRetries, err = kpr.fetchDKGParamsForKeyperSet(ctx, ev.Contract)
		if err != nil {
			return errors.Wrap(err, "fetch DKG params for new keyper set")
		}
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
			if err := coredb.InsertEon(ctx, corekeyperdb.InsertEonParams{
				Eon:                   keyperSetIndex,
				ActivationBlockNumber: activationBlockNumber,
				KeyperConfigIndex:     keyperSetIndex,
				DkgContract:           dkgContract,
				PhaseLength:           phaseLength,
				LeadLength:            leadLength,
				MaxRetries:            maxRetries,
			}); err != nil {
				return errors.Wrap(err, "insert eon row for new keyper set")
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Must run after the transaction commits: MaybeRegisterECIESKey reads
	// the keyper set row inserted above.
	//
	// This can still land too late. If DKG starts immediately (the
	// KeyperSetAdded tx was mined inside the lead-time window), peers
	// dispatch dealings before our ECIES registration is on chain and
	// encrypt without our key. Returning keypers already have a key from
	// a previous set, so this only bites new members. Mitigation: give
	// keyper sets enough lead time.
	return kpr.dkgMgr.MaybeRegisterECIESKey(ctx, keyperSetIndex)
}

// fetchDKGParamsForKeyperSet asks the keyper set contract for its DKG contract
// address and then reads the immutable phase parameters from that DKG contract.
// Any failure (zero address, RPC error, missing methods on an old contract) is
// returned as an error. The DKG manager treats NULL phase params in the eons
// row as a fatal configuration error and has no chain-client fallback, so we
// must not insert the eon row without these values; propagating the error lets
// the outer event loop log and retry on the next chainsync re-delivery.
func (kpr *Keyper) fetchDKGParamsForKeyperSet(
	ctx context.Context,
	keyperSetAddr common.Address,
) (sql.NullString, sql.NullInt64, sql.NullInt64, int64, error) {
	var (
		nullStr sql.NullString
		nullInt sql.NullInt64
	)
	if (keyperSetAddr == common.Address{}) {
		return nullStr, nullInt, nullInt, 0, errors.New("keyper set event missing contract address")
	}
	ks, err := keypersetBindings.NewKeyperset(keyperSetAddr, kpr.chainSyncClient.Client)
	if err != nil {
		return nullStr, nullInt, nullInt, 0, errors.Wrapf(err, "bind keyper set contract %s", keyperSetAddr.Hex())
	}
	callOpts := &bind.CallOpts{Context: ctx}
	dkgAddr, err := ks.GetDKGContract(callOpts)
	if err != nil {
		return nullStr, nullInt, nullInt, 0, errors.Wrapf(err, "call getDKGContract on keyper set %s", keyperSetAddr.Hex())
	}
	if (dkgAddr == common.Address{}) {
		return nullStr, nullInt, nullInt, 0, errors.Errorf("keyper set %s has no DKG contract configured", keyperSetAddr.Hex())
	}
	dkg, err := dkgcontract.NewDkgcontract(dkgAddr, kpr.chainSyncClient.Client)
	if err != nil {
		return nullStr, nullInt, nullInt, 0, errors.Wrapf(err, "bind DKG contract %s", dkgAddr.Hex())
	}
	phaseLength, err := dkg.PHASELENGTH(callOpts)
	if err != nil {
		return nullStr, nullInt, nullInt, 0, errors.Wrapf(err, "read PHASE_LENGTH from %s", dkgAddr.Hex())
	}
	leadLength, err := dkg.DKGLEADLENGTH(callOpts)
	if err != nil {
		return nullStr, nullInt, nullInt, 0, errors.Wrapf(err, "read DKG_LEAD_LENGTH from %s", dkgAddr.Hex())
	}
	maxRetries, err := dkg.MAXRETRIES(callOpts)
	if err != nil {
		return nullStr, nullInt, nullInt, 0, errors.Wrapf(err, "read MAX_RETRIES from %s", dkgAddr.Hex())
	}
	log.Info().
		Str("keyper-set", keyperSetAddr.Hex()).
		Str("dkg-contract", dkgAddr.Hex()).
		Uint64("phase-length", phaseLength).
		Uint64("lead-length", leadLength).
		Uint64("max-retries", maxRetries).
		Msg("resolved per-keyper-set DKG contract params")
	//nolint:gosec // G115: phase length, lead length, and max retries come from the on-chain contract and fit well within int64
	return sql.NullString{String: dkgAddr.Hex(), Valid: true},
		sql.NullInt64{Int64: int64(phaseLength), Valid: true},
		sql.NullInt64{Int64: int64(leadLength), Valid: true},
		int64(maxRetries),
		nil
}
