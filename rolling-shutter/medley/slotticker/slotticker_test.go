package slotticker

import (
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestCalcNextSlot(t *testing.T) {
	duration := time.Second * 5
	genesisTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	epsilon := time.Millisecond

	for _, testCase := range []struct {
		timeSinceGenesis time.Duration
		offset           time.Duration
		slot             uint64
	}{
		{0, 0, 0},
		{-time.Second * 100, 0, 0},
		{epsilon, 0, 1},
		{duration / 2, 0, 1},
		{duration, 0, 1},
		{2 * duration, 0, 2},
		{100*duration - epsilon, 0, 100},

		{-time.Second, -time.Second, 0},
		{0, -time.Second, 1},
		{4 * time.Second, -time.Second, 1},
		{4*time.Second + epsilon, -time.Second, 2},
		{100*duration - time.Second, -time.Second, 100},
		{100*duration - time.Second + epsilon, -time.Second, 101},
	} {
		t.Run("", func(t *testing.T) {
			now := genesisTime.Add(testCase.timeSinceGenesis)
			slot, tick := calcNextTick(now, genesisTime, duration, testCase.offset)
			assert.Equal(t, testCase.slot, slot)
			expectedTick := genesisTime.Add(duration * time.Duration(testCase.slot)).Add(testCase.offset)
			assert.Equal(t, tick, expectedTick)
		})
	}
}
