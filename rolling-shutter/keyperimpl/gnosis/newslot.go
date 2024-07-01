package gnosis

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	gnosisdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/gnosis/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/slotticker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func (kpr *Keyper) processNewSlot(ctx context.Context, slot slotticker.Slot) error {
	return kpr.maybeTriggerDecryption(ctx, slot.Number)
}

// maybeTriggerDecryption triggers decryption for the given slot if
// - it hasn't been triggered for this slot before and
// - the keyper is part of the corresponding keyper set.
func (kpr *Keyper) maybeTriggerDecryption(ctx context.Context, slot uint64) error {
	if kpr.latestTriggeredSlot != nil && slot <= *kpr.latestTriggeredSlot {
		return nil
	}
	kpr.latestTriggeredSlot = &slot

	fmt.Println("")
	fmt.Println("")
	fmt.Println(slot)
	fmt.Println("")
	fmt.Println("")

	gnosisKeyperDB := gnosisdatabase.New(kpr.dbpool)
	syncedUntil, err := gnosisKeyperDB.GetTransactionSubmittedEventsSyncedUntil(ctx)
	if err != nil && err != pgx.ErrNoRows {
		// pgx.ErrNoRows is expected if we're not part of the keyper set (which is checked later).
		// That's because non-keypers don't sync transaction submitted events.
		return errors.Wrap(err, "failed to query transaction submitted sync status from db")
	}
	if syncedUntil.Slot >= int64(slot) {
		// If we already synced the block for slot n before this slot has started on our clock,
		// either the previous block proposer proposed early (ie is malicious) or our clocks are
		// out of sync. In any case, it does not make sense to produce keys as the block has
		// already been built, so we return an error.
		return errors.Errorf("processing slot %d for which a block has already been processed", slot)
	}
	nextBlock := syncedUntil.BlockNumber + 1

	queries := obskeyper.New(kpr.dbpool)
	keyperSet, err := queries.GetKeyperSet(ctx, nextBlock)
	if err == pgx.ErrNoRows {
		log.Debug().
			Uint64("slot", slot).
			Int64("block-number", nextBlock).
			Msg("skipping slot as no keyper set has been found for it")
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to query keyper set for block %d", nextBlock)
	}

	// don't trigger if we're not part of the keyper set
	if !keyperSet.Contains(kpr.config.GetAddress()) {
		log.Debug().
			Uint64("slot", slot).
			Int64("block-number", nextBlock).
			Int64("keyper-set-index", keyperSet.KeyperConfigIndex).
			Str("address", kpr.config.GetAddress().Hex()).
			Msg("skipping slot as not part of keyper set")
		return nil
	}

	// don't trigger if the block proposer is not part of the validator registry
	isRegistered, proposerIndex, err := kpr.isProposerRegistered(ctx, slot, uint64(nextBlock))
	if err != nil {
		return err
	}
	if !isRegistered {
		log.Debug().
			Uint64("slot", slot).
			Uint64("proposer-index", proposerIndex).
			Msg("skipping slot as proposer is not registered")
		return nil
	}

	// For each block with a proposer, the tx pointer age has to be incremented. If keys for
	// this slot are successfully produced, it will be reset to 0. If for some reason the key
	// has already been produced and the tx pointer has been subsequently reset (e.g. because
	// we are lagging behind), the tx pointer age will end up being 1. Being off by 1 doesn't
	// matter much though because it will only become relevant once it reaches maxTxPointerAge
	// which is arbitrary and locally chosen anyway.
	age, err := gnosisKeyperDB.IncrementTxPointerAge(ctx, keyperSet.KeyperConfigIndex)
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "failed to increment tx pointer age")
	}
	if age.Valid {
		log.Debug().
			Uint64("slot", slot).
			Int64("block-number", nextBlock).
			Int64("keyper-config-index", keyperSet.KeyperConfigIndex).
			Int64("new-tx-pointer-age", age.Int64).
			Msg("incremented tx pointer age")
	} else {
		log.Warn().
			Uint64("slot", slot).
			Int64("block-number", nextBlock).
			Int64("keyper-config-index", keyperSet.KeyperConfigIndex).
			Msg("tx pointer age is infinite")
	}

	return kpr.triggerDecryption(ctx, slot, nextBlock, &keyperSet)
}

func (kpr *Keyper) isProposerRegistered(ctx context.Context, slot uint64, block uint64) (bool, uint64, error) {
	epoch := medley.SlotToEpoch(slot, kpr.config.Gnosis.SlotsPerEpoch)
	proposerDuties, err := kpr.beaconAPIClient.GetProposerDutiesByEpoch(ctx, epoch)
	if err != nil {
		return false, 0, err
	}
	if proposerDuties == nil {
		return false, 0, errors.Errorf("no proposer duties found for slot %d in epoch %d", slot, epoch)
	}
	proposerDuty, err := proposerDuties.GetDutyForSlot(slot)
	if err != nil {
		return false, 0, err
	}
	proposerIndex := proposerDuty.ValidatorIndex
	if proposerIndex > math.MaxInt64 {
		return false, 0, errors.New("proposer index too big")
	}

	db := gnosisdatabase.New(kpr.dbpool)
	isRegistered, err := db.IsValidatorRegistered(ctx, gnosisdatabase.IsValidatorRegisteredParams{
		ValidatorIndex: int64(proposerDuty.ValidatorIndex),
		BlockNumber:    int64(block),
	})
	if err == pgx.ErrNoRows {
		return false, proposerIndex, nil
	}
	if err != nil {
		return false, 0, errors.Wrapf(err, "failed to query registration status for validator %d", proposerDuty.ValidatorIndex)
	}
	return isRegistered, proposerDuty.ValidatorIndex, nil
}

func getTxPointer(ctx context.Context, db *pgxpool.Pool, eon int64, maxTxPointerAge int64) (int64, error) {
	gnosisKeyperDB := gnosisdatabase.New(db)
	var txPointer int64
	var txPointerAge int64
	var txPointerOutdated bool
	eonString := fmt.Sprint(eon)
	txPointerDB, err := gnosisKeyperDB.GetTxPointer(ctx, eon)
	if err == pgx.ErrNoRows {
		log.Info().Int64("eon", eon).Msg("initializing tx pointer")
		// If there is no tx pointer in the db, we initialize it to 0. This is the intended case
		// for newly started eons. We might also end up doing that if the keyper has been started
		// late for an eon. In this case, we would ideally set it to infinity (as we would do on a
		// restart). 0 is ok though too.
		err = gnosisKeyperDB.SetTxPointer(ctx, gnosisdatabase.SetTxPointerParams{
			Eon: eon,
			Age: sql.NullInt64{
				Int64: 0,
				Valid: true,
			},
			Value: 0,
		})
		if err != nil {
			return 0, errors.Wrapf(err, "failed to initialize tx pointer for eon %d", eon)
		}
		txPointer = 0
		txPointerAge = 0
		txPointerOutdated = false
	} else if err != nil {
		return 0, errors.Wrap(err, "failed to query tx pointer from db")
	} else {
		txPointer = txPointerDB.Value
		txPointerAge = txPointerDB.Age.Int64
		if txPointerDB.Age.Valid {
			txPointerOutdated = txPointerAge > maxTxPointerAge
		} else {
			txPointerAge = math.MaxInt64
			txPointerOutdated = true
		}
	}

	// If the tx pointer is outdated, the system has failed to generate decryption keys (or at
	// least we haven't received them). This either means not enough keypers are online or they
	// don't agree on the current value of the tx pointer. In order to recover, we choose the
	// current length of the transaction queue as the new tx pointer, as this is a value
	// everyone can agree on.
	if txPointerOutdated {
		log.Warn().
			Int64("eon", eon).
			Int64("tx-pointer", txPointer).
			Int64("tx-pointer-age", txPointerAge).
			Msg("outdated tx pointer")
		txPointer, err = gnosisKeyperDB.GetTransactionSubmittedEventCount(ctx, eon)
		if err == pgx.ErrNoRows {
			txPointer = 0
		} else if err != nil {
			return 0, errors.Wrap(err, "failed to query transaction submitted event count from db")
		}
	}
	metricsTxPointer.WithLabelValues(eonString).Set(float64(txPointer))
	metricsTxPointerAge.WithLabelValues(eonString).Set(float64(txPointerAge))
	return txPointer, nil
}

func (kpr *Keyper) triggerDecryption(
	ctx context.Context,
	slot uint64,
	nextBlock int64,
	keyperSet *obskeyper.KeyperSet,
) error {
	gnosisKeyperDB := gnosisdatabase.New(kpr.dbpool)
	coreKeyperDB := corekeyperdatabase.New(kpr.dbpool)

	eonStruct, err := coreKeyperDB.GetEonForBlockNumber(ctx, nextBlock)
	if err != nil {
		return errors.Wrapf(err, "failed to query eon for block number %d from db", nextBlock)
	}
	keyperConfigIndex := eonStruct.KeyperConfigIndex

	txPointer, err := getTxPointer(ctx, kpr.dbpool, keyperConfigIndex, int64(kpr.config.Gnosis.MaxTxPointerAge))
	if err != nil {
		return err
	}

	identityPreimages, err := kpr.getDecryptionIdentityPreimages(ctx, slot, keyperSet.KeyperConfigIndex, txPointer)
	if err != nil {
		return err
	}
	err = gnosisKeyperDB.SetCurrentDecryptionTrigger(ctx, gnosisdatabase.SetCurrentDecryptionTriggerParams{
		Eon:            keyperConfigIndex,
		Slot:           int64(slot),
		TxPointer:      txPointer,
		IdentitiesHash: computeIdentitiesHash(identityPreimages),
	})
	if err != nil {
		return errors.Wrap(err, "failed to insert published tx pointer into db")
	}
	trigger := epochkghandler.DecryptionTrigger{
		BlockNumber:       uint64(nextBlock),
		IdentityPreimages: identityPreimages,
	}
	event := broker.NewEvent(&trigger)
	log.Debug().
		Uint64("slot", slot).
		Uint64("block-number", uint64(nextBlock)).
		Int("num-identities", len(trigger.IdentityPreimages)).
		Int64("tx-pointer", txPointer).
		Msg("sending decryption trigger")
	kpr.decryptionTriggerChannel <- event

	return nil
}

func (kpr *Keyper) getDecryptionIdentityPreimages(
	ctx context.Context, slot uint64, eon int64, txPointer int64,
) ([]identitypreimage.IdentityPreimage, error) {
	identityPreimages := []identitypreimage.IdentityPreimage{}

	queries := gnosisdatabase.New(kpr.dbpool)
	limitUint64 := kpr.config.Gnosis.EncryptedGasLimit/kpr.config.Gnosis.MinGasPerTransaction + 1
	if limitUint64 > math.MaxInt32 {
		return identityPreimages, errors.New("gas limit too big")
	}
	limit := int32(limitUint64)

	events, err := queries.GetTransactionSubmittedEvents(ctx, gnosisdatabase.GetTransactionSubmittedEventsParams{
		Eon:   eon,
		Index: txPointer,
		Limit: limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query transaction submitted events from index %d", txPointer)
	}

	identityPreimages = []identitypreimage.IdentityPreimage{
		makeSlotIdentityPreimage(slot),
	}
	gas := uint64(0)
	for _, event := range events {
		gas += uint64(event.GasLimit)
		if gas > kpr.config.Gnosis.EncryptedGasLimit {
			break
		}
		identityPreimage, err := transactionSubmittedEventToIdentityPreimage(event)
		if err != nil {
			return []identitypreimage.IdentityPreimage{}, err
		}
		identityPreimages = append(identityPreimages, identityPreimage)
	}

	sortedIdentityPreimages := sortIdentityPreimages(identityPreimages)

	return sortedIdentityPreimages, nil
}

func transactionSubmittedEventToIdentityPreimage(
	event gnosisdatabase.TransactionSubmittedEvent,
) (identitypreimage.IdentityPreimage, error) {
	sender, err := shdb.DecodeAddress(event.Sender)
	if err != nil {
		return identitypreimage.IdentityPreimage{}, errors.Wrap(err, "failed to decode sender address of transaction submitted event from db")
	}

	var buf bytes.Buffer
	buf.Write(event.IdentityPrefix)
	buf.Write(sender.Bytes())

	return identitypreimage.IdentityPreimage(buf.Bytes()), nil
}

func makeSlotIdentityPreimage(slot uint64) identitypreimage.IdentityPreimage {
	// 32 bytes of zeros plus the block number as 20 byte big endian (ie starting with lots of
	// zeros as well). This ensures the block identity preimage is always alphanumerically before
	// any transaction identity preimages, because sender addresses cannot be that small.
	var buf bytes.Buffer
	buf.Write(common.BigToHash(common.Big0).Bytes())
	buf.Write(common.BigToHash(new(big.Int).SetUint64(slot)).Bytes()[12:])

	return identitypreimage.IdentityPreimage(buf.Bytes())
}

func sortIdentityPreimages(identityPreimages []identitypreimage.IdentityPreimage) []identitypreimage.IdentityPreimage {
	sorted := make([]identitypreimage.IdentityPreimage, len(identityPreimages))
	copy(sorted, identityPreimages)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i], sorted[j]) < 0
	})
	return sorted
}
