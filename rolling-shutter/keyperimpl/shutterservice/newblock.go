package shutterservice

import (
	"bytes"
	"context"
	"encoding/hex"
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
	if kpr.config.EventBasedTriggersEnabled() {
		err := kpr.multiEventSyncer.Sync(ctx, ev.Header)
		if err != nil {
			return err
		}
	}
	return kpr.maybeTriggerDecryption(ctx, ev)
}

func (kpr *Keyper) maybeTriggerDecryption(ctx context.Context, block *syncevent.LatestBlock) error {
	timeBasedTriggers, err := kpr.prepareTimeBasedTriggers(ctx, block)
	if err != nil {
		return errors.Wrap(err, "failed to get time based triggers")
	}
	kpr.sendTriggers(ctx, timeBasedTriggers)

	if kpr.config.EventBasedTriggersEnabled() {
		eventBasedTriggers, err := kpr.prepareEventBasedTriggers(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get event based triggers")
		}
		kpr.sendTriggers(ctx, eventBasedTriggers)
	}

	return nil
}

func (kpr *Keyper) prepareTimeBasedTriggers(ctx context.Context, block *syncevent.LatestBlock) ([]epochkghandler.DecryptionTrigger, error) {
	if kpr.latestTriggeredTime != nil && block.Header.Time <= *kpr.latestTriggeredTime {
		return nil, nil
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
		return nil, errors.Wrap(err, "failed to query non decrypted identity registered events from db")
	}

	eventsToDecrypt := make([]servicedatabase.IdentityRegisteredEvent, 0)
	for _, event := range nonTriggeredEvents {
		trigger, err := kpr.shouldTriggerDecryption(ctx, event, block)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to check if should trigger decryption for event %d", event.Eon)
		}
		if trigger {
			eventsToDecrypt = append(eventsToDecrypt, event)
		}
	}

	return kpr.createTriggersFromIdentityRegisteredEvents(ctx, eventsToDecrypt, block)
}

func (kpr *Keyper) shouldTriggerDecryption(
	ctx context.Context,
	event servicedatabase.IdentityRegisteredEvent,
	triggeredBlock *syncevent.LatestBlock,
) (bool, error) {
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
			return false, nil
		} else {
			return false, errors.Wrapf(err, "failed to query keyper state for eon %d", event.Eon)
		}
	}

	eon, err := coreKeyperDB.GetEon(ctx, event.Eon)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get eon %d", event.Eon)
	}
	if eon.ActivationBlockNumber > triggeredBlock.Header.Number.Int64() {
		log.Info().
			Int64("eon", event.Eon).
			Int64("block-number", triggeredBlock.Header.Number.Int64()).
			Msg("skipping event as eon activation block number is greater than triggered block number")
		return false, nil
	}

	if event.Timestamp >= int64(triggeredBlock.Header.Time) {
		return false, nil
	}

	// don't trigger if we're not part of the keyper set
	if !isKeyper {
		log.Info().
			Int64("eon", event.Eon).
			Int64("block-number", event.BlockNumber).
			Str("identity", hex.EncodeToString(event.Identity)).
			Str("address", kpr.config.GetAddress().Hex()).
			Msg("skipping event as not part of keyper set")
		return false, nil
	}
	return true, nil
}

func (kpr *Keyper) createTriggersFromIdentityRegisteredEvents(
	ctx context.Context,
	triggeredEvents []servicedatabase.IdentityRegisteredEvent,
	triggeredBlock *syncevent.LatestBlock,
) ([]epochkghandler.DecryptionTrigger, error) {
	coreKeyperDB := corekeyperdatabase.New(kpr.dbpool)
	identityPreimages := make(map[int64][]identitypreimage.IdentityPreimage)
	lastEonBlock := make(map[int64]int64)
	for _, event := range triggeredEvents {
		eon, err := coreKeyperDB.GetEon(ctx, event.Eon)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to query eon %d from db", event.Eon)
		}

		if identityPreimages[event.Eon] == nil {
			identityPreimages[event.Eon] = make([]identitypreimage.IdentityPreimage, 0)
		}
		identityPreimages[event.Eon] = append(identityPreimages[event.Eon], identitypreimage.IdentityPreimage(event.Identity))

		if _, exists := lastEonBlock[event.Eon]; !exists {
			lastEonBlock[event.Eon] = eon.ActivationBlockNumber
		}
	}

	triggers := []epochkghandler.DecryptionTrigger{}
	for eon, preImages := range identityPreimages {
		sortedIdentityPreimages := sortIdentityPreimages(preImages)

		trigger := epochkghandler.DecryptionTrigger{
			// sending last block available for that eon as the key shares will be generated based on the eon associated with this block number
			BlockNumber:       uint64(lastEonBlock[eon]),
			IdentityPreimages: sortedIdentityPreimages,
		}
		triggers = append(triggers, trigger)
	}
	return triggers, nil
}

func (kpr *Keyper) prepareEventBasedTriggers(ctx context.Context) ([]epochkghandler.DecryptionTrigger, error) {
	coreKeyperDB := corekeyperdatabase.New(kpr.dbpool)
	serviceDB := servicedatabase.New(kpr.dbpool)
	obsDB := obskeyper.New(kpr.dbpool)

	firedTriggers, err := serviceDB.GetUndecryptedFiredTriggers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get undecrypted fired triggers from db")
	}

	firedTriggersByEon := make(map[int64][]servicedatabase.GetUndecryptedFiredTriggersRow)
	for _, firedTrigger := range firedTriggers {
		firedTriggersByEon[firedTrigger.Eon] = append(firedTriggersByEon[firedTrigger.Eon], firedTrigger)
	}

	var decryptionTriggers []epochkghandler.DecryptionTrigger
	for eon, firedTriggers := range firedTriggersByEon {
		if len(firedTriggers) == 0 {
			continue
		}
		eonStruct, err := coreKeyperDB.GetEon(ctx, eon)
		if err != nil {
			if err == pgx.ErrNoRows {
				log.Info().
					Int64("eon", eon).
					Msg("ignoring fired triggers as eon not found in db")
				continue
			}
			return nil, errors.Wrapf(err, "failed to query eon %d from db", eon)
		}
		keyperSet, err := obsDB.GetKeyperSet(ctx, eonStruct.ActivationBlockNumber)
		if err != nil {
			log.Err(err).
				Int64("eon", eon).
				Int64("activation-block-number", eonStruct.ActivationBlockNumber).
				Msg("ignoring fired triggers as keyper set not found in db")
			continue
		}
		if !keyperSet.Contains(kpr.config.GetAddress()) {
			log.Info().
				Int64("eon", eon).
				Int64("activation-block-number", eonStruct.ActivationBlockNumber).
				Str("address", kpr.config.GetAddress().Hex()).
				Msg("ignoring fired triggers as not part of keyper set")
			continue
		}

		identities := []identitypreimage.IdentityPreimage{}
		for _, firedTrigger := range firedTriggers {
			identities = append(identities, firedTrigger.Identity)
		}

		sortedIdentityPreimages := sortIdentityPreimages(identities)

		decryptionTrigger := epochkghandler.DecryptionTrigger{
			BlockNumber:       uint64(eonStruct.ActivationBlockNumber),
			IdentityPreimages: sortedIdentityPreimages,
		}

		decryptionTriggers = append(decryptionTriggers, decryptionTrigger)
	}
	return decryptionTriggers, nil
}

func (kpr *Keyper) sendTriggers(ctx context.Context, triggers []epochkghandler.DecryptionTrigger) {
	for _, trigger := range triggers {
		event := broker.NewEvent(&trigger)
		log.Debug().
			Uint64("eon", trigger.BlockNumber).
			Int("num-identities", len(trigger.IdentityPreimages)).
			Msg("sending decryption trigger")

		select {
		case kpr.decryptionTriggerChannel <- event:
		case <-ctx.Done():
			log.Warn().
				Err(ctx.Err()).
				Msg("context canceled while sending decryption trigger")
			return
		}
	}
}

func sortIdentityPreimages(identityPreimages []identitypreimage.IdentityPreimage) []identitypreimage.IdentityPreimage {
	sorted := make([]identitypreimage.IdentityPreimage, len(identityPreimages))
	copy(sorted, identityPreimages)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i], sorted[j]) < 0
	})
	return sorted
}
