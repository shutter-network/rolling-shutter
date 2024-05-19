package gnosis

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
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

// Maximum age of a tx pointer in blocks before it is considered outdated.
const maxTxPointerAge = 2

var errZeroTxPointerAge = errors.New("tx pointer has age 0")

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
	if err != nil {
		return errors.Wrap(err, "failed to query synced until from db")
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
	isRegistered, err := kpr.isProposerRegistered(ctx, slot)
	if err != nil {
		return err
	}
	if !isRegistered {
		log.Debug().
			Uint64("slot", slot).
			Msg("skipping slot as proposer is not registered")
		// Even if we don't trigger decryption, we still need to update the tx pointer or it will
		// become outdated.
		err := gnosisKeyperDB.SetTxPointerSlot(ctx, gnosisdatabase.SetTxPointerSlotParams{
			Eon:  keyperSet.KeyperConfigIndex,
			Slot: int64(slot),
		})
		if err != nil {
			return errors.Wrap(err, "failed to update tx pointer slot")
		}
		return nil
	}

	return kpr.triggerDecryption(ctx, slot, nextBlock, &keyperSet)
}

func (kpr *Keyper) isProposerRegistered(ctx context.Context, slot uint64) (bool, error) {
	epoch := medley.SlotToEpoch(slot, kpr.config.Gnosis.SlotsPerEpoch)
	proposerDuties, err := kpr.beaconAPIClient.GetProposerDutiesByEpoch(ctx, epoch)
	if err != nil {
		return false, err
	}
	if proposerDuties == nil {
		return false, errors.Errorf("no proposer duties found for slot %d in epoch %d", slot, epoch)
	}
	proposerDuty, err := proposerDuties.GetDutyForSlot(slot)
	if err != nil {
		return false, err
	}
	proposerIndex := proposerDuty.ValidatorIndex
	if proposerIndex > math.MaxInt64 {
		return false, errors.New("proposer index too big")
	}

	db := gnosisdatabase.New(kpr.dbpool)
	isRegistered, err := db.IsValidatorRegistered(ctx, gnosisdatabase.IsValidatorRegisteredParams{
		ValidatorIndex: int64(proposerDuty.ValidatorIndex),
		BlockNumber:    int64(block),
	})
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrapf(err, "failed to query registration status for validator %d", proposerDuty.ValidatorIndex)
	}
	return isRegistered, nil
}

func (kpr *Keyper) getTxPointer(ctx context.Context, eon int64, slot int64, keyperConfigIndex int64) (int64, error) {
	gnosisKeyperDB := gnosisdatabase.New(kpr.dbpool)
	var txPointer, txPointerAge int64
	txPointerDB, err := gnosisKeyperDB.GetTxPointer(ctx, eon)
	if err == pgx.ErrNoRows {
		// The tx pointer is expected to be missing from the db if the eon has just started. In
		// this case, we should initialize it to zero with an age of 1, ie decrypt starting with
		// the first transaction.
		// The tx pointer may also be missing if the keyper has been started late and no decryption
		// key has been generated or received yet (receiving the keys message would update the
		// pointer). In this case, the true age is unknown, as we only know the start block but
		// not the start slot of the eon. However, we can ignore this edge case as it will be
		// resolved automatically when the first keys message is received. If key generation
		// continues to fail, eventually our tx pointer age will exceed the maximum value and we
		// will start participating in the recovery process, albeit a bit late.
		err := gnosisKeyperDB.SetTxPointer(ctx, gnosisdatabase.SetTxPointerParams{
			Eon:   eon,
			Slot:  slot,
			Value: 0,
		})
		if err != nil {
			return 0, errors.Wrap(err, "failed to initialize tx pointer")
		}
		txPointer = 0
		txPointerAge = 1
	} else if err != nil {
		return 0, errors.Wrap(err, "failed to query tx pointer from db")
	} else {
		txPointer = txPointerDB.Value
		txPointerAge = slot - txPointerDB.Slot
	}
	if txPointerAge == 0 {
		// A pointer of age 0 means we already received the pointer from a DecryptionKeys message
		// even though we haven't sent our shares yet. In that case, sending our shares is
		// unnecessary.
		return 0, errZeroTxPointerAge
	}
	// If the tx pointer is outdated, the system has failed to generate decryption keys (or at
	// least we haven't received them). This either means not enough keypers are online or they
	// don't agree on the current value of the tx pointer. In order to recover, we choose the
	// current length of the transaction queue as the new tx pointer, as this is a value
	// everyone can agree on.
	isOutdated := txPointerAge > maxTxPointerAge
	if isOutdated {
		log.Warn().
			Int64("slot", slot).
			Int64("eon", eon).
			Int64("tx-pointer", txPointer).
			Int64("tx-pointer-age", txPointerAge).
			Msg("outdated tx pointer")
		txPointer, err = gnosisKeyperDB.GetTransactionSubmittedEventCount(ctx, keyperConfigIndex)
		if err == pgx.ErrNoRows {
			txPointer = 0
		} else if err != nil {
			return 0, errors.Wrap(err, "failed to query transaction submitted event count from db")
		}
	}
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

	txPointer, err := kpr.getTxPointer(ctx, keyperConfigIndex, int64(slot), keyperSet.KeyperConfigIndex)
	if err == errZeroTxPointerAge {
		log.Warn().
			Uint64("slot", slot).
			Int64("block-number", nextBlock).
			Int64("eon", keyperConfigIndex).
			Int64("tx-pointer", txPointer).
			Msg("skipping trigger as tx pointer age is 0")
		return nil
	} else if err != nil {
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
	return identityPreimages, nil
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
	// 32 bytes of zeros plus the block number as big endian (ie starting with lots of zeros as well)
	// this ensures the block identity preimage is always alphanumerically before any transaction
	// identity preimages.
	var buf bytes.Buffer
	buf.Write(common.BigToHash(common.Big0).Bytes())
	buf.Write(common.BigToHash(new(big.Int).SetUint64(slot)).Bytes())

	return identitypreimage.IdentityPreimage(buf.Bytes())
}
