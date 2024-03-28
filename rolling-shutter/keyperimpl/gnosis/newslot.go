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
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/slotticker"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/shdb"
)

func (kpr *Keyper) processNewSlot(ctx context.Context, slot slotticker.Slot) error {
	gnosisKeyperDB := gnosisdatabase.New(kpr.dbpool)
	syncedUntil, err := gnosisKeyperDB.GetTransactionSubmittedEventsSyncedUntil(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to query synced until from db")
	}
	if syncedUntil.Slot >= int64(slot.Number) {
		// If we already synced the block for slot n before this slot has started on our clock,
		// either the previous block proposer proposed early (ie is malicious) or our clocks are
		// out of sync. In any case, it does not make sense to produce keys as the block has
		// already been built, so we return an error.
		return errors.Errorf("processing slot %d for which a block has already been processed", slot.Number)
	}

	queries := obskeyper.New(kpr.dbpool)
	keyperSet, err := queries.GetKeyperSet(ctx, syncedUntil.BlockNumber)
	if err == pgx.ErrNoRows {
		log.Debug().
			Uint64("slot", slot.Number).
			Int64("block-number", syncedUntil.BlockNumber).
			Msg("ignoring slot as no keyper set has been found for it")
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to query keyper set for block %d", syncedUntil.BlockNumber)
	}
	for _, m := range keyperSet.Keypers {
		if m == shdb.EncodeAddress(kpr.config.GetAddress()) {
			return kpr.triggerDecryption(ctx, slot, syncedUntil, &keyperSet)
		}
	}
	log.Debug().Uint64("slot", slot.Number).Msg("ignoring block as not part of keyper set")
	return nil
}

func (kpr *Keyper) triggerDecryption(
	ctx context.Context,
	slot slotticker.Slot,
	syncedUntil gnosisdatabase.TransactionSubmittedEventsSyncedUntil,
	keyperSet *obskeyper.KeyperSet,
) error {
	fmt.Println("")
	fmt.Println("")
	fmt.Println(slot.Number)
	fmt.Println("")
	fmt.Println("")
	gnosisKeyperDB := gnosisdatabase.New(kpr.dbpool)
	coreKeyperDB := corekeyperdatabase.New(kpr.dbpool)

	eonStruct, err := coreKeyperDB.GetEonForBlockNumber(ctx, syncedUntil.BlockNumber)
	if err != nil {
		return errors.Wrapf(err, "failed to query eon for block number %d from db", syncedUntil.BlockNumber)
	}
	eon := eonStruct.Eon

	var txPointer int64
	var txPointerAge int64
	txPointerDB, err := gnosisKeyperDB.GetTxPointer(ctx, eon)
	if err == pgx.ErrNoRows {
		txPointer = 0
		txPointerAge = syncedUntil.BlockNumber - keyperSet.ActivationBlockNumber + 1
	} else if err != nil {
		return errors.Wrap(err, "failed to query tx pointer from db")
	} else {
		txPointerAge = syncedUntil.BlockNumber - txPointerDB.Block
		txPointer = txPointerDB.Value
	}
	if txPointerAge == 0 {
		// A pointer of age 0 means we already received the pointer from a DecryptionKeys message
		// even though we haven't sent our shares yet. In that case, sending our shares is
		// unnecessary.
		log.Warn().
			Uint64("slot", slot.Number).
			Int64("block-number", syncedUntil.BlockNumber).
			Int64("eon", eon).
			Int64("tx-pointer", txPointer).
			Int64("tx-pointer-age", txPointerAge).
			Msg("ignoring new block as tx pointer age is 0")
		return nil
	}
	if txPointerAge > maxTxPointerAge {
		// If the tx pointer is outdated, the system has failed to generate decryption keys (or at
		// least we haven't received them). This either means not enough keypers are online or they
		// don't agree on the current value of the tx pointer. In order to recover, we choose the
		// current length of the transaction queue as the new tx pointer, as this is a value
		// everyone can agree on.
		log.Warn().
			Uint64("slot", slot.Number).
			Int64("block-number", syncedUntil.BlockNumber).
			Int64("eon", eon).
			Int64("tx-pointer", txPointer).
			Int64("tx-pointer-age", txPointerAge).
			Msg("outdated tx pointer")
		txPointer, err = gnosisKeyperDB.GetTransactionSubmittedEventCount(ctx, keyperSet.KeyperConfigIndex)
		if err == pgx.ErrNoRows {
			txPointer = 0
		} else if err != nil {
			return errors.Wrap(err, "failed to query transaction submitted event count from db")
		}
	}

	identityPreimages, err := kpr.getDecryptionIdentityPreimages(ctx, slot, keyperSet.KeyperConfigIndex, txPointer)
	if err != nil {
		return err
	}
	err = gnosisKeyperDB.SetCurrentDecryptionTrigger(ctx, gnosisdatabase.SetCurrentDecryptionTriggerParams{
		Eon:            eon,
		Slot:           int64(slot.Number),
		TxPointer:      txPointer,
		IdentitiesHash: computeIdentitiesHash(identityPreimages),
	})
	if err != nil {
		return errors.Wrap(err, "failed to insert published tx pointer into db")
	}
	trigger := epochkghandler.DecryptionTrigger{
		BlockNumber:       uint64(syncedUntil.BlockNumber),
		IdentityPreimages: identityPreimages,
	}
	event := broker.NewEvent(&trigger)
	log.Debug().
		Uint64("slot", slot.Number).
		Uint64("block-number", uint64(syncedUntil.BlockNumber)).
		Int("num-identities", len(trigger.IdentityPreimages)).
		Int64("tx-pointer", txPointer).
		Int64("tx-pointer-age", txPointerAge).
		Msg("sending decryption trigger")
	kpr.decryptionTriggerChannel <- event

	return nil
}

func (kpr *Keyper) getDecryptionIdentityPreimages(
	ctx context.Context, slot slotticker.Slot, eon int64, txPointer int64,
) ([]identitypreimage.IdentityPreimage, error) {
	identityPreimages := []identitypreimage.IdentityPreimage{}

	queries := gnosisdatabase.New(kpr.dbpool)
	limitUint64 := kpr.config.EncryptedGasLimit/kpr.config.MinGasPerTransaction + 1
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
		if gas > kpr.config.EncryptedGasLimit {
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

func makeSlotIdentityPreimage(slot slotticker.Slot) identitypreimage.IdentityPreimage {
	// 32 bytes of zeros plus the block number as big endian (ie starting with lots of zeros as well)
	// this ensures the block identity preimage is always alphanumerically before any transaction
	// identity preimages.
	var buf bytes.Buffer
	buf.Write(common.BigToHash(common.Big0).Bytes())
	buf.Write(common.BigToHash(new(big.Int).SetUint64(slot.Number)).Bytes())

	return identitypreimage.IdentityPreimage(buf.Bytes())
}
