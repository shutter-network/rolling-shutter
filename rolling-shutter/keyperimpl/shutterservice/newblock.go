package shutterservice

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

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
	kpr.latestTriggeredTime = &block.Header.Time

	fmt.Println("--------------")
	fmt.Println(block.Header.Time)
	fmt.Println("--------------")

	serviceDB := servicedatabase.New(kpr.dbpool)
	nonTriggeredEvents, err := serviceDB.GetNotDecryptedIdentityRegisteredEvents(ctx,
		servicedatabase.GetNotDecryptedIdentityRegisteredEventsParams{
			Timestamp:   int64(lastTriggeredTime),
			Timestamp_2: int64(block.Header.Time),
		})
	if err != nil && err != pgx.ErrNoRows {
		// pgx.ErrNoRows is expected if we're not part of the keyper set (which is checked later).
		// That's because non-keypers don't sync identity registered events. TODO: this needs to be implemented
		return errors.Wrap(err, "failed to query non decrypted identity registered events from db")
	}

	eventsToDecrypt := make([]servicedatabase.IdentityRegisteredEvent, 0)
	for _, event := range nonTriggeredEvents {
		if kpr.shouldTriggerDecryption(ctx, event, block) {
			eventsToDecrypt = append(eventsToDecrypt, event)
		}
	}

	return kpr.triggerDecryption(ctx, eventsToDecrypt, block)
}

func (kpr *Keyper) shouldTriggerDecryption(
	ctx context.Context,
	event servicedatabase.IdentityRegisteredEvent,
	triggeredBlock *syncevent.LatestBlock,
) bool {
	coreKeyperDB := corekeyperdatabase.New(kpr.dbpool)
	isKeyper, err := coreKeyperDB.GetKeyperStateForEon(ctx, corekeyperdatabase.GetKeyperStateForEonParams{
		Eon:           event.Eon,
		KeyperAddress: []string{kpr.config.GetAddress().Hex()},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Info().
				Int64("eon", event.Eon).
				Msg("skipping event as no eon has been found for it")
		} else {
			log.Err(err).Msgf("failed to query keyper state for eon %d", event.Eon)
		}
		return false
	}

	eon, err := coreKeyperDB.GetEon(ctx, event.Eon)
	if err != nil {
		log.Err(err).Msgf("failed to get eon %d", event.Eon)
		return false
	}
	if eon.ActivationBlockNumber > triggeredBlock.Header.Number.Int64() {
		log.Info().
			Int64("eon", event.Eon).
			Int64("block-number", triggeredBlock.Header.Number.Int64()).
			Msg("skipping event as eon activation block number is greater than triggered block number")
		return false
	}

	if event.Timestamp >= int64(triggeredBlock.Header.Time) {
		return false
	}

	// don't trigger if we're not part of the keyper set
	if !isKeyper {
		log.Info().
			Int64("eon", event.Eon).
			Int64("block-number", event.BlockNumber).
			Str("identity", string(event.Identity)).
			Str("address", kpr.config.GetAddress().Hex()).
			Msg("skipping event as not part of keyper set")
		return false
	}
	return true
}

func (kpr *Keyper) triggerDecryption(ctx context.Context,
	triggeredEvents []servicedatabase.IdentityRegisteredEvent,
	triggeredBlock *syncevent.LatestBlock,
) error {
	coreKeyperDB := corekeyperdatabase.New(kpr.dbpool)
	identityPreimages := make(map[int64][]identitypreimage.IdentityPreimage)
	lastEonBlock := make(map[int64]int64)
	for _, event := range triggeredEvents {
		eon, err := coreKeyperDB.GetEon(ctx, event.Eon)
		if err != nil {
			return errors.Wrap(err, "failed to get eon")
		}

		if identityPreimages[event.Eon] == nil {
			identityPreimages[event.Eon] = make([]identitypreimage.IdentityPreimage, 0)
		}
		identityPreimages[event.Eon] = append(identityPreimages[event.Eon], identitypreimage.IdentityPreimage(event.Identity))

		if _, exists := lastEonBlock[event.Eon]; !exists {
			lastEonBlock[event.Eon] = eon.ActivationBlockNumber
		}
	}

	for eon, preImages := range identityPreimages {
		sortedIdentityPreimages := sortIdentityPreimages(preImages)

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
