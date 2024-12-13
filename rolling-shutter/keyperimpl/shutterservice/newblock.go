package shutterservice

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	obskeyper "github.com/shutter-network/rolling-shutter/rolling-shutter/chainobserver/db/keyper"
	corekeyperdatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyper/epochkghandler"
	servicedatabase "github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/database"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/broker"
	syncevent "github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
)

func (kpr *Keyper) processNewBlock(ctx context.Context, ev *syncevent.LatestBlock) error {
	if kpr.registrySyncer != nil {
		if err := kpr.registrySyncer.Sync(ctx, ev.Header); err != nil {
			return err
		}
	}
	return kpr.maybeTriggerDecryption(ctx, ev)
}

// maybeTriggerDecryption triggers decryption for the identities registered if
// - it hasn't been triggered for thos identities before and
// - the keyper is part of the corresponding keyper set.
func (kpr *Keyper) maybeTriggerDecryption(ctx context.Context, block *syncevent.LatestBlock) error {
	if kpr.latestTriggeredTime != nil && block.Header.Time <= *kpr.latestTriggeredTime {
		return nil
	}

	lastTriggeredTime := 0
	if kpr.latestTriggeredTime != nil {
		lastTriggeredTime = int(*kpr.latestTriggeredTime)
	}

	fmt.Println("--------------")
	fmt.Println(block.Header.Time)
	fmt.Println("--------------")

	serviceDB := servicedatabase.New(kpr.dbpool)
	nonTriggeredEvents, err := serviceDB.GetNotDecryptedIdentityRegisteredEvents(ctx, servicedatabase.GetNotDecryptedIdentityRegisteredEventsParams{
		Timestamp:   int64(lastTriggeredTime),
		Timestamp_2: int64(block.Header.Time),
	})
	if err != nil && err != pgx.ErrNoRows {
		// pgx.ErrNoRows is expected if we're not part of the keyper set (which is checked later).
		// That's because non-keypers don't sync identity registered events. TODO: this needs to be implemented
		return errors.Wrap(err, "failed to query non decrypted identity registered events from db")
	}

	obsDB := obskeyper.New(kpr.dbpool)
	eventsToDecrypt := make([]servicedatabase.IdentityRegisteredEvent, 0)
	for _, event := range nonTriggeredEvents {
		if kpr.shouldTriggerDecryption(ctx, obsDB, &event, block) {
			eventsToDecrypt = append(eventsToDecrypt, event)
		}
	}

	err = kpr.triggerDecryption(ctx, eventsToDecrypt, block)
	if err != nil {
		return err
	}
	kpr.latestTriggeredTime = &block.Header.Time
	return nil
}

func (kpr *Keyper) shouldTriggerDecryption(
	ctx context.Context,
	obsDB *obskeyper.Queries,
	event *servicedatabase.IdentityRegisteredEvent,
	triggeredBlock *syncevent.LatestBlock,
) bool {
	nextBlock := triggeredBlock.Number.Int64()
	keyperSet, err := obsDB.GetKeyperSet(ctx, nextBlock)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Info().
				Int64("block-number", nextBlock).
				Msg("skipping event as no keyper set has been found for it")
		} else {
			log.Err(err).Msgf("failed to query keyper set for block %d", nextBlock)
		}
		return false
	}

	if event.Timestamp >= int64(triggeredBlock.Header.Time) {
		return false
	}

	// don't trigger if we're not part of the keyper set
	if !keyperSet.Contains(kpr.config.GetAddress()) {
		log.Info().
			Int64("block-number", nextBlock).
			Int64("keyper-set-index", keyperSet.KeyperConfigIndex).
			Str("address", kpr.config.GetAddress().Hex()).
			Msg("skipping slot as not part of keyper set")
		return false
	}
	return true
}

func (kpr *Keyper) triggerDecryption(ctx context.Context,
	triggeredEvents []servicedatabase.IdentityRegisteredEvent,
	triggeredBlock *syncevent.LatestBlock,
) error {
	coreKeyperDB := corekeyperdatabase.New(kpr.dbpool)
	serviceDB := servicedatabase.New(kpr.dbpool)

	identityPreimages := make(map[int64][]identitypreimage.IdentityPreimage)
	lastEonBlock := make(map[int64]int64)
	for _, event := range triggeredEvents {
		nextBlock := triggeredBlock.Header.Number.Int64()

		eonStruct, err := coreKeyperDB.GetEonForBlockNumber(ctx, nextBlock)
		if err != nil {
			return errors.Wrapf(err, "failed to query eon for block number %d from db", nextBlock)
		}

		if eonStruct.Eon != event.Eon {
			log.Warn().
				Int64("eon expected", eonStruct.Eon).
				Int64("eon in event", event.Eon).
				Msg("skipping event as wrong eon passed")

			continue
		}

		if identityPreimages[event.Eon] == nil {
			identityPreimages[event.Eon] = make([]identitypreimage.IdentityPreimage, 0)
		}
		identityPreimages[event.Eon] = append(identityPreimages[event.Eon], identitypreimage.IdentityPreimage(event.Identity))

		if lastEonBlock[event.Eon] < event.BlockNumber {
			lastEonBlock[event.Eon] = event.BlockNumber
		}
	}

	for eon, preImages := range identityPreimages {
		sortedIdentityPreimages := sortIdentityPreimages(preImages)

		err := serviceDB.SetCurrentDecryptionTrigger(ctx, servicedatabase.SetCurrentDecryptionTriggerParams{
			Eon:                  eon,
			TriggeredBlockNumber: triggeredBlock.Number.Int64(),
			IdentitiesHash:       computeIdentitiesHash(sortedIdentityPreimages),
		})
		if err != nil {
			return errors.Wrap(err, "failed to insert published decryption trigger into db")
		}

		trigger := epochkghandler.DecryptionTrigger{
			// sending last block available for that eon as the key shares will be generated based on the eon associated with this block number
			BlockNumber:       uint64(lastEonBlock[eon]),
			IdentityPreimages: sortedIdentityPreimages,
		}

		event := broker.NewEvent(&trigger)
		log.Debug().
			Uint64("block-number", uint64(lastEonBlock[eon])).
			Int("num-identities", len(trigger.IdentityPreimages)).
			Msg("sending decryption trigger")
		kpr.decryptionTriggerChannel <- event
	}

	return nil
}

func sortIdentityPreimages(identityPreimages []identitypreimage.IdentityPreimage) []identitypreimage.IdentityPreimage {
	sorted := make([]identitypreimage.IdentityPreimage, len(identityPreimages))
	copy(sorted, identityPreimages)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i], sorted[j]) < 0
	})
	return sorted
}
