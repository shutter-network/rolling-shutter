package syncer

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/shutter-network/contracts/v2/bindings/dkgcontract"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/logger"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

// DKGContractSyncer must be an independently startable service.Service, and
// fakeDKGBackend must satisfy the leaf's backend interface.
var (
	_ service.Service = (*DKGContractSyncer)(nil)
	_ dkgEventWatcher = (*fakeDKGBackend)(nil)
)

func blockNumber(u uint64) *number.BlockNumber {
	return number.NewBlockNumber(&u)
}

// newContractSyncer builds a DKGContractSyncer wired to a fake backend, with a
// recording handler so tests can pull delivered events off a channel.
func newContractSyncer(addr common.Address, backend dkgEventWatcher, startBlock uint64) (*DKGContractSyncer, *recordingHandler) {
	handler := newRecordingHandler()
	s := &DKGContractSyncer{
		Addr:       addr,
		Log:        &logger.NoopLogger{},
		Handler:    handler.Handle,
		StartBlock: blockNumber(startBlock),
		backend:    backend,
	}
	return s, handler
}

func TestDKGContractSyncerStartRequiresHandler(t *testing.T) {
	s := &DKGContractSyncer{
		Log:        &logger.NoopLogger{},
		StartBlock: blockNumber(1),
		backend:    newFakeDKGBackend(common.HexToAddress("0xaa")),
	}
	runner := newFakeRunner(context.Background())
	t.Cleanup(runner.Wait)
	err := s.Start(context.Background(), runner)
	assert.Assert(t, err != nil, "Start without a handler must error")
}

func TestDKGContractSyncerDeliversDealingEvent(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	s, handler := newContractSyncer(addr, backend, 42)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	backend.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 7, KeyperIndex: 3})

	ev := handler.expectNextEvent(t)
	de, ok := ev.(*event.DealingEvent)
	assert.Assert(t, ok, "expected *DealingEvent, got %T", ev)
	assert.Equal(t, uint64(7), de.KeyperSetIndex)
	assert.Equal(t, uint64(3), de.KeyperIndex)
}

func TestDKGContractSyncerDeliversAccusationEvent(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	s, handler := newContractSyncer(addr, backend, 42)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	backend.emitAccusation(t, &dkgcontract.DkgcontractAccusationSubmitted{KeyperSetIndex: 1, KeyperIndex: 2})

	ev := handler.expectNextEvent(t)
	ae, ok := ev.(*event.AccusationEvent)
	assert.Assert(t, ok, "expected *AccusationEvent, got %T", ev)
	assert.Equal(t, uint64(1), ae.KeyperSetIndex)
	assert.Equal(t, uint64(2), ae.KeyperIndex)
}

func TestDKGContractSyncerDeliversApologyEvent(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	s, handler := newContractSyncer(addr, backend, 42)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	backend.emitApology(t, &dkgcontract.DkgcontractApologySubmitted{KeyperSetIndex: 4, KeyperIndex: 5})

	ev := handler.expectNextEvent(t)
	ae, ok := ev.(*event.ApologyEvent)
	assert.Assert(t, ok, "expected *ApologyEvent, got %T", ev)
	assert.Equal(t, uint64(4), ae.KeyperSetIndex)
	assert.Equal(t, uint64(5), ae.KeyperIndex)
}

func TestDKGContractSyncerDeliversSuccessVoteEvent(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	s, handler := newContractSyncer(addr, backend, 42)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	backend.emitSuccessVote(t, &dkgcontract.DkgcontractSuccessVoteSubmitted{KeyperSetIndex: 8, KeyperIndex: 9})

	ev := handler.expectNextEvent(t)
	sve, ok := ev.(*event.SuccessVoteEvent)
	assert.Assert(t, ok, "expected *SuccessVoteEvent, got %T", ev)
	assert.Equal(t, uint64(8), sve.KeyperSetIndex)
	assert.Equal(t, uint64(9), sve.KeyperIndex)
}

func TestDKGContractSyncerDeliversSuccessEvent(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	s, handler := newContractSyncer(addr, backend, 42)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	backend.emitSuccess(t, &dkgcontract.DkgcontractDKGSucceeded{KeyperSetIndex: 6, RetryCounter: 2})

	ev := handler.expectNextEvent(t)
	se, ok := ev.(*event.SuccessEvent)
	assert.Assert(t, ok, "expected *SuccessEvent, got %T", ev)
	assert.Equal(t, uint64(6), se.KeyperSetIndex)
	assert.Equal(t, uint64(2), se.RetryCounter)
}

func TestDKGContractSyncerWatchStartIsRequestedBlock(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	s, _ := newContractSyncer(addr, backend, 4242)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	assert.Equal(t, uint64(4242), backend.dealingStart, "DealingSubmitted watch must start from requested block")
	assert.Equal(t, uint64(4242), backend.accusationStart)
	assert.Equal(t, uint64(4242), backend.apologyStart)
	assert.Equal(t, uint64(4242), backend.successVoteStart)
	assert.Equal(t, uint64(4242), backend.successStart)
}

func TestDKGContractSyncerSubscriptionSetupErrorPropagates(t *testing.T) {
	for _, eventType := range []string{"dealing", "accusation", "apology", "successVote", "success"} {
		t.Run(eventType, func(t *testing.T) {
			addr := common.HexToAddress("0xaa")
			backend := newFakeDKGBackend(addr)
			backend.watchErrs = map[string]error{eventType: errors.New("simulated subscription failure")}
			s, _ := newContractSyncer(addr, backend, 1)

			runner := newFakeRunner(context.Background())
			t.Cleanup(runner.Wait)

			err := s.Start(context.Background(), runner)
			assert.Assert(t, err != nil, "subscription setup error on %s channel must propagate from Start", eventType)
		})
	}
}

func TestDKGContractSyncerContextCancellationExitsCleanly(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	s, _ := newContractSyncer(addr, backend, 1)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)

	assert.NilError(t, s.Start(ctx, runner))

	cancel()

	done := make(chan struct{})
	go func() {
		runner.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("watch goroutine did not exit after context cancellation")
	}
}
