package slotticker

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type Slot struct {
	Number          uint64
	genesisSlotTime time.Time
	slotDuration    time.Duration
}

func (s Slot) Start() time.Time {
	return s.genesisSlotTime.Add(s.slotDuration * time.Duration(s.Number))
}

// SlotTicker is a ticker that ticks at the start of each slot.
type SlotTicker struct {
	C               chan Slot
	slotDuration    time.Duration
	genesisSlotTime time.Time
	offset          time.Duration
}

func NewSlotTicker(slotDuration time.Duration, genesisSlotTime time.Time, offset time.Duration) *SlotTicker {
	c := make(chan Slot, 1)
	return &SlotTicker{
		C:               c,
		slotDuration:    slotDuration,
		genesisSlotTime: genesisSlotTime,
		offset:          offset,
	}
}

func (t *SlotTicker) tick(ctx context.Context, n uint64) error {
	s := Slot{
		Number:          n,
		genesisSlotTime: t.genesisSlotTime,
		slotDuration:    t.slotDuration,
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case t.C <- s:
		return nil
	}
}

//nolint:unparam
func (t *SlotTicker) Start(ctx context.Context, runner service.Runner) error {
	runner.Go(func() error {
		return t.run(ctx)
	})
	return nil
}

func (t *SlotTicker) run(ctx context.Context) error {
	var prevSlotNumber *uint64 = nil
	timer := time.NewTimer(0)
	<-timer.C

	for {
		now := time.Now()
		nextSlotNumber, nextTickTime := calcNextTick(now, t.genesisSlotTime, t.slotDuration, t.offset)

		if prevSlotNumber != nil {
			expectedNextSlotNumber := *prevSlotNumber + 1
			if nextSlotNumber < expectedNextSlotNumber {
				// This should never happen unless the system clock changes. If it does, there's
				// nothing we can do about it.
				log.Error().
					Uint64("next-slot-number", nextSlotNumber).
					Uint64("prev-slot-number", *prevSlotNumber).
					Msg("slot ticker emitted slots in wrong order")
			} else if nextSlotNumber > expectedNextSlotNumber {
				log.Warn().
					Uint64("next-slot-number", nextSlotNumber).
					Uint64("prev-slot-number", *prevSlotNumber).
					Msg("missing slots due to slow slot processing")
				for i := expectedNextSlotNumber; i < nextSlotNumber; i++ {
					if err := t.tick(ctx, i); err != nil {
						return err
					}
				}
			}
		}

		timeToNextSlot := nextTickTime.Sub(now)
		timer.Reset(timeToNextSlot)
		<-timer.C

		if err := t.tick(ctx, nextSlotNumber); err != nil {
			return err
		}

		prevSlotNumber = &nextSlotNumber
	}
}

func calcNextTick(now time.Time, genesisSlotTime time.Time, slotDuration time.Duration, offset time.Duration) (uint64, time.Time) {
	firstTick := genesisSlotTime.Add(offset)
	if now.Before(firstTick) {
		return 0, firstTick
	}
	slot := uint64((now.Sub(genesisSlotTime) - offset + slotDuration - 1) / slotDuration)
	tick := genesisSlotTime.Add(slotDuration * time.Duration(slot)).Add(offset)
	return slot, tick
}
