package syncer

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethevent "github.com/ethereum/go-ethereum/event"
	gethlog "github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/logger"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

// recordingLogger captures every Warn call for assertion in tests. All other
// log levels are inherited from NoopLogger.
type recordingLogger struct {
	*logger.NoopLogger
	mu    sync.Mutex
	warns []logEntry
}

type logEntry struct {
	msg string
	ctx []interface{}
}

func newRecordingLogger() *recordingLogger {
	return &recordingLogger{NoopLogger: &logger.NoopLogger{}}
}

func (r *recordingLogger) Warn(msg string, ctx ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.warns = append(r.warns, logEntry{msg: msg, ctx: ctx})
}

func (r *recordingLogger) warnCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.warns)
}

// fakeKeyperSetIndexer implements keyperSetIndexer for unit tests.
type fakeKeyperSetIndexer struct {
	addresses []common.Address
}

func (f *fakeKeyperSetIndexer) GetNumKeyperSets(*bind.CallOpts) (uint64, error) {
	return uint64(len(f.addresses)), nil
}

func (f *fakeKeyperSetIndexer) GetKeyperSetAddress(_ *bind.CallOpts, index uint64) (common.Address, error) {
	if index >= uint64(len(f.addresses)) {
		return common.Address{}, errors.New("index out of range")
	}
	return f.addresses[index], nil
}

// stubResolver returns a resolver that maps each KeyperSet address to a
// preconfigured DKG contract address.
func stubResolver(mapping map[common.Address]common.Address) dkgContractResolver {
	return func(_ context.Context, _ *bind.CallOpts, ksAddr common.Address) (common.Address, error) {
		addr, ok := mapping[ksAddr]
		if !ok {
			return common.Address{}, errors.Errorf("unknown keyper set %s", ksAddr.Hex())
		}
		return addr, nil
	}
}

func newSyncer(t *testing.T, resolver dkgContractResolver) (*DKGEventSyncer, *recordingLogger) {
	t.Helper()
	rec := newRecordingLogger()
	s := &DKGEventSyncer{
		Log:                rec,
		resolveDKGContract: resolver,
	}
	return s, rec
}

func TestRecordDKGContractAddsNewAddress(t *testing.T) {
	s, rec := newSyncer(t, nil)
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")

	added := s.recordDKGContract(addr, 0, common.HexToAddress("0xaa"))
	assert.Equal(t, true, added, "first insertion should be reported as newly added")
	assert.Equal(t, 0, rec.warnCount(), "non-zero address must not produce a warning")
	assert.DeepEqual(t, []common.Address{addr}, s.trackedDKGContractList())
}

func TestRecordDKGContractDeduplicates(t *testing.T) {
	s, _ := newSyncer(t, nil)
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")

	addedFirst := s.recordDKGContract(addr, 0, common.HexToAddress("0xaa"))
	addedSecond := s.recordDKGContract(addr, 1, common.HexToAddress("0xbb"))

	assert.Equal(t, true, addedFirst)
	assert.Equal(t, false, addedSecond, "second insertion of the same address must be reported as not newly added")
	assert.Equal(t, 1, len(s.trackedDKGContractList()), "tracked set must contain a single entry after dedup")
}

func TestRecordDKGContractZeroAddressWarnsAndSkips(t *testing.T) {
	s, rec := newSyncer(t, nil)

	added := s.recordDKGContract(common.Address{}, 7, common.HexToAddress("0xaa"))

	assert.Equal(t, false, added, "zero address must not be added")
	assert.Equal(t, 0, len(s.trackedDKGContractList()))
	assert.Equal(t, 1, rec.warnCount(), "zero address must produce exactly one warning")
}

func TestScanInitialDKGContractsPopulatesSet(t *testing.T) {
	ks0 := common.HexToAddress("0xaa")
	ks1 := common.HexToAddress("0xbb")
	dkg0 := common.HexToAddress("0xcc")
	dkg1 := common.HexToAddress("0xdd")

	indexer := &fakeKeyperSetIndexer{addresses: []common.Address{ks0, ks1}}
	resolver := stubResolver(map[common.Address]common.Address{ks0: dkg0, ks1: dkg1})
	s, _ := newSyncer(t, resolver)

	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	err := s.scanInitialDKGContracts(context.Background(), opts, indexer)
	assert.NilError(t, err)

	got := s.trackedDKGContractList()
	assert.Equal(t, 2, len(got))
	want := map[common.Address]struct{}{dkg0: {}, dkg1: {}}
	for _, addr := range got {
		_, ok := want[addr]
		assert.Equal(t, true, ok, "unexpected DKG contract %s in tracked set", addr.Hex())
	}
}

func TestScanInitialDKGContractsDeduplicatesSharedContract(t *testing.T) {
	ks0 := common.HexToAddress("0xaa")
	ks1 := common.HexToAddress("0xbb")
	ks2 := common.HexToAddress("0xcc")
	shared := common.HexToAddress("0xdd")

	indexer := &fakeKeyperSetIndexer{addresses: []common.Address{ks0, ks1, ks2}}
	resolver := stubResolver(map[common.Address]common.Address{ks0: shared, ks1: shared, ks2: shared})
	s, _ := newSyncer(t, resolver)

	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	err := s.scanInitialDKGContracts(context.Background(), opts, indexer)
	assert.NilError(t, err)

	assert.DeepEqual(t, []common.Address{shared}, s.trackedDKGContractList())
}

func TestScanInitialDKGContractsZeroAddressSkippedWithWarning(t *testing.T) {
	ks0 := common.HexToAddress("0xaa")
	ks1 := common.HexToAddress("0xbb")
	dkg1 := common.HexToAddress("0xdd")

	indexer := &fakeKeyperSetIndexer{addresses: []common.Address{ks0, ks1}}
	resolver := stubResolver(map[common.Address]common.Address{ks0: {}, ks1: dkg1})
	s, rec := newSyncer(t, resolver)

	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	err := s.scanInitialDKGContracts(context.Background(), opts, indexer)
	assert.NilError(t, err)

	assert.DeepEqual(t, []common.Address{dkg1}, s.trackedDKGContractList())
	assert.Equal(t, 1, rec.warnCount(), "exactly one warning should be emitted for the zero-address keyper set")
}

func TestScanInitialDKGContractsEmptyKeyperSetManager(t *testing.T) {
	indexer := &fakeKeyperSetIndexer{}
	resolver := stubResolver(nil)
	s, rec := newSyncer(t, resolver)

	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	err := s.scanInitialDKGContracts(context.Background(), opts, indexer)
	assert.NilError(t, err)

	assert.Equal(t, 0, len(s.trackedDKGContractList()))
	assert.Equal(t, 0, rec.warnCount())
}

func TestHandleKeyperSetAddedRecordsNewContract(t *testing.T) {
	ksAddr := common.HexToAddress("0xaa")
	dkgAddr := common.HexToAddress("0xbb")
	resolver := stubResolver(map[common.Address]common.Address{ksAddr: dkgAddr})
	s, _ := newSyncer(t, resolver)

	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ksAddr,
		Eon:               3,
	}
	ev.Raw.BlockNumber = 42

	s.handleKeyperSetAdded(context.Background(), ev)

	assert.DeepEqual(t, []common.Address{dkgAddr}, s.trackedDKGContractList())
}

func TestHandleKeyperSetAddedDeduplicatesAgainstExisting(t *testing.T) {
	ks0 := common.HexToAddress("0xaa")
	ks1 := common.HexToAddress("0xbb")
	shared := common.HexToAddress("0xcc")
	resolver := stubResolver(map[common.Address]common.Address{ks0: shared, ks1: shared})
	s, _ := newSyncer(t, resolver)

	indexer := &fakeKeyperSetIndexer{addresses: []common.Address{ks0}}
	opts := &bind.CallOpts{Context: context.Background(), BlockNumber: big.NewInt(1)}
	assert.NilError(t, s.scanInitialDKGContracts(context.Background(), opts, indexer))
	assert.DeepEqual(t, []common.Address{shared}, s.trackedDKGContractList())

	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ks1,
		Eon:               1,
	}
	ev.Raw.BlockNumber = 99
	s.handleKeyperSetAdded(context.Background(), ev)

	assert.DeepEqual(t, []common.Address{shared}, s.trackedDKGContractList())
}

func TestHandleKeyperSetAddedZeroAddressLogsWarning(t *testing.T) {
	ksAddr := common.HexToAddress("0xaa")
	resolver := stubResolver(map[common.Address]common.Address{ksAddr: {}})
	s, rec := newSyncer(t, resolver)

	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ksAddr,
		Eon:               5,
	}
	s.handleKeyperSetAdded(context.Background(), ev)

	assert.Equal(t, 0, len(s.trackedDKGContractList()))
	assert.Equal(t, 1, rec.warnCount())
}

func TestHandleKeyperSetAddedResolverErrorIsToleratedWithoutTracking(t *testing.T) {
	resolver := func(context.Context, *bind.CallOpts, common.Address) (common.Address, error) {
		return common.Address{}, errors.New("simulated RPC failure")
	}
	s, _ := newSyncer(t, resolver)

	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: common.HexToAddress("0xaa"),
		Eon:               9,
	}
	s.handleKeyperSetAdded(context.Background(), ev)

	assert.Equal(t, 0, len(s.trackedDKGContractList()),
		"a resolver error must not result in any DKG contract being tracked")
}

// gethLoggerInterface compile-time assertion: recordingLogger must satisfy the
// log.Logger interface so it can be wired into DKGEventSyncer.Log in real
// scenarios as well as in these tests.
var _ gethlog.Logger = (*recordingLogger)(nil)

// fakeSubscription is a no-op gethevent.Subscription whose Err channel never
// fires and Unsubscribe is idempotent. Tests that need subscription errors
// to surface drive them via a separate channel.
type fakeSubscription struct {
	errCh chan error
	once  sync.Once
}

func newFakeSubscription() *fakeSubscription {
	return &fakeSubscription{errCh: make(chan error, 1)}
}

func (f *fakeSubscription) Err() <-chan error { return f.errCh }
func (f *fakeSubscription) Unsubscribe()      { f.once.Do(func() { close(f.errCh) }) }

// fakeDKGBackend is an in-memory dkgContractBackend that records the watch
// start block per event type and lets tests fire events by calling the
// emit* helpers. Subscription channels are captured on the first Watch call;
// tests should set up the syncer before emitting events.
type fakeDKGBackend struct {
	addr      common.Address
	succeeded map[uint64]bool

	mu              sync.Mutex
	dealingSink     chan<- *contract.DKGContractDealingSubmitted
	accusationSink  chan<- *contract.DKGContractAccusationSubmitted
	apologySink     chan<- *contract.DKGContractApologySubmitted
	successVoteSink chan<- *contract.DKGContractSuccessVoteSubmitted
	successSink     chan<- *contract.DKGContractDKGSucceeded

	dealingStart     uint64
	accusationStart  uint64
	apologyStart     uint64
	successVoteStart uint64
	successStart     uint64
}

func newFakeDKGBackend(addr common.Address) *fakeDKGBackend {
	return &fakeDKGBackend{addr: addr, succeeded: map[uint64]bool{}}
}

func (b *fakeDKGBackend) Succeeded(_ *bind.CallOpts, ksi uint64) (bool, error) {
	return b.succeeded[ksi], nil
}

func (b *fakeDKGBackend) WatchDealingSubmitted(
	opts *bind.WatchOpts, sink chan<- *contract.DKGContractDealingSubmitted, _, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dealingSink = sink
	if opts != nil && opts.Start != nil {
		b.dealingStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) WatchAccusationSubmitted(
	opts *bind.WatchOpts, sink chan<- *contract.DKGContractAccusationSubmitted, _, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.accusationSink = sink
	if opts != nil && opts.Start != nil {
		b.accusationStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) WatchApologySubmitted(
	opts *bind.WatchOpts, sink chan<- *contract.DKGContractApologySubmitted, _, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.apologySink = sink
	if opts != nil && opts.Start != nil {
		b.apologyStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) WatchSuccessVoteSubmitted(
	opts *bind.WatchOpts, sink chan<- *contract.DKGContractSuccessVoteSubmitted, _, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.successVoteSink = sink
	if opts != nil && opts.Start != nil {
		b.successVoteStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) WatchDKGSucceeded(
	opts *bind.WatchOpts, sink chan<- *contract.DKGContractDKGSucceeded, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.successSink = sink
	if opts != nil && opts.Start != nil {
		b.successStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) emitDealing(t *testing.T, ev *contract.DKGContractDealingSubmitted) {
	t.Helper()
	b.mu.Lock()
	sink := b.dealingSink
	b.mu.Unlock()
	assert.Assert(t, sink != nil, "dealing subscription not set up yet")
	select {
	case sink <- ev:
	case <-time.After(time.Second):
		t.Fatal("dealing channel blocked")
	}
}

func (b *fakeDKGBackend) emitSuccess(t *testing.T, ev *contract.DKGContractDKGSucceeded) {
	t.Helper()
	b.mu.Lock()
	sink := b.successSink
	b.mu.Unlock()
	assert.Assert(t, sink != nil, "success subscription not set up yet")
	select {
	case sink <- ev:
	case <-time.After(time.Second):
		t.Fatal("success channel blocked")
	}
}

// fakeRunner implements service.Runner for tests. Goroutines spawned via Go
// are tracked so the test can wait for them to drain after cancelling its
// context.
type fakeRunner struct {
	wg sync.WaitGroup
}

func newFakeRunner() *fakeRunner { return &fakeRunner{} }

func (r *fakeRunner) Go(f func() error) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		_ = f()
	}()
}

func (r *fakeRunner) StartService(_ ...service.Service) error { return nil }

func (r *fakeRunner) Defer(_ func()) {}

func (r *fakeRunner) Wait() { r.wg.Wait() }

// recordingHandler delivers every DKG event the syncer hands it onto a
// buffered channel so tests can pull them off in order.
type recordingHandler struct {
	ch chan event.DKGEvent
}

func newRecordingHandler() *recordingHandler {
	return &recordingHandler{ch: make(chan event.DKGEvent, 32)}
}

func (h *recordingHandler) Handle(_ context.Context, ev event.DKGEvent) error {
	h.ch <- ev
	return nil
}

// expectNextEvent waits up to the timeout for one more event and returns it.
func (h *recordingHandler) expectNextEvent(t *testing.T) event.DKGEvent {
	t.Helper()
	select {
	case ev := <-h.ch:
		return ev
	case <-time.After(2 * time.Second):
		t.Fatal("expected DKG event, got nothing")
		return nil
	}
}

func newSubscriptionSyncer(t *testing.T, backends map[common.Address]*fakeDKGBackend) (*DKGEventSyncer, *recordingHandler) {
	t.Helper()
	handler := newRecordingHandler()
	binder := func(addr common.Address) (dkgContractBackend, error) {
		b, ok := backends[addr]
		if !ok {
			return nil, errors.Errorf("unknown DKG contract %s", addr.Hex())
		}
		return b, nil
	}
	s := &DKGEventSyncer{
		Log:             &logger.NoopLogger{},
		Handler:         handler.Handle,
		bindDKGContract: binder,
	}
	return s, handler
}

func TestStartContractSubscriptionDeliversLiveDealingEvent(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	s, handler := newSubscriptionSyncer(t, map[common.Address]*fakeDKGBackend{addr: backend})

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner()
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.startContractSubscription(ctx, runner, addr, 42))

	backend.emitDealing(t, &contract.DKGContractDealingSubmitted{
		KeyperSetIndex: 7,
		KeyperIndex:    3,
	})

	ev := handler.expectNextEvent(t)
	de, ok := ev.(*event.DealingEvent)
	assert.Assert(t, ok, "expected *DealingEvent, got %T", ev)
	assert.Equal(t, uint64(7), de.KeyperSetIndex)
	assert.Equal(t, uint64(3), de.KeyperIndex)
}

func TestStartContractSubscriptionWatchStartIsRequestedBlock(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	s, _ := newSubscriptionSyncer(t, map[common.Address]*fakeDKGBackend{addr: backend})

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner()
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.startContractSubscription(ctx, runner, addr, 4242))

	assert.Equal(t, uint64(4242), backend.dealingStart, "DealingSubmitted watch must start from requested block")
	assert.Equal(t, uint64(4242), backend.accusationStart)
	assert.Equal(t, uint64(4242), backend.apologyStart)
	assert.Equal(t, uint64(4242), backend.successVoteStart)
	assert.Equal(t, uint64(4242), backend.successStart)
}

func TestStartContractSubscriptionInitialSuccessesDeliveredBeforeLiveEvents(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	backend.succeeded[0] = true
	backend.succeeded[2] = true

	s, handler := newSubscriptionSyncer(t, map[common.Address]*fakeDKGBackend{addr: backend})
	s.updateKnownKeyperSetCount(3)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner()
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.startContractSubscription(ctx, runner, addr, 100))

	// Initial successes are delivered synchronously before startContractSubscription
	// returns, so reading from the handler channel first must yield SuccessEvents.
	ev1 := handler.expectNextEvent(t)
	_, ok := ev1.(*event.SuccessEvent)
	assert.Assert(t, ok, "first delivered event must be initial *SuccessEvent, got %T", ev1)

	ev2 := handler.expectNextEvent(t)
	_, ok = ev2.(*event.SuccessEvent)
	assert.Assert(t, ok, "second delivered event must be initial *SuccessEvent, got %T", ev2)

	// Only after both initial successes is the live event emitted; it must
	// arrive after them on the same channel.
	backend.emitDealing(t, &contract.DKGContractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: 1})
	ev3 := handler.expectNextEvent(t)
	_, ok = ev3.(*event.DealingEvent)
	assert.Assert(t, ok, "live event after initial successes must be *DealingEvent, got %T", ev3)
}

func TestStartViaTrackedListDeliversEventsFromTwoDistinctContracts(t *testing.T) {
	addr0 := common.HexToAddress("0xa0")
	addr1 := common.HexToAddress("0xa1")
	backend0 := newFakeDKGBackend(addr0)
	backend1 := newFakeDKGBackend(addr1)

	s, handler := newSubscriptionSyncer(t, map[common.Address]*fakeDKGBackend{
		addr0: backend0,
		addr1: backend1,
	})
	// Both addresses are pre-tracked, simulating the outcome of the initial
	// scan -- this is the same state Start() leaves the syncer in before it
	// spawns per-contract subscriptions.
	s.recordDKGContract(addr0, 0, common.HexToAddress("0xb0"))
	s.recordDKGContract(addr1, 1, common.HexToAddress("0xb1"))

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner()
	t.Cleanup(func() { cancel(); runner.Wait() })

	for _, addr := range s.trackedDKGContractList() {
		assert.NilError(t, s.startContractSubscription(ctx, runner, addr, 100))
	}

	backend0.emitDealing(t, &contract.DKGContractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: 1})
	backend1.emitSuccess(t, &contract.DKGContractDKGSucceeded{KeyperSetIndex: 1, RetryCounter: 4})

	got := map[uint64]bool{}
	for i := 0; i < 2; i++ {
		ev := handler.expectNextEvent(t)
		switch e := ev.(type) {
		case *event.DealingEvent:
			got[e.KeyperSetIndex] = true
		case *event.SuccessEvent:
			got[e.KeyperSetIndex] = true
		default:
			t.Fatalf("unexpected event type %T", ev)
		}
	}
	assert.Assert(t, got[0], "DealingEvent from contract 0 was not delivered")
	assert.Assert(t, got[1], "SuccessEvent from contract 1 was not delivered")
}

func TestSharedContractIsSubscribedExactlyOnce(t *testing.T) {
	shared := common.HexToAddress("0xa0")
	backend := newFakeDKGBackend(shared)
	s, handler := newSubscriptionSyncer(t, map[common.Address]*fakeDKGBackend{shared: backend})
	s.recordDKGContract(shared, 0, common.HexToAddress("0xb0"))
	s.recordDKGContract(shared, 1, common.HexToAddress("0xb1"))
	s.recordDKGContract(shared, 2, common.HexToAddress("0xb2"))

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner()
	t.Cleanup(func() { cancel(); runner.Wait() })

	// trackedDKGContractList must contain exactly one entry after dedup.
	tracked := s.trackedDKGContractList()
	assert.Equal(t, 1, len(tracked), "shared DKG contract must appear in tracked list exactly once")
	for _, addr := range tracked {
		assert.NilError(t, s.startContractSubscription(ctx, runner, addr, 100))
	}

	backend.emitDealing(t, &contract.DKGContractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: 5})

	// First delivery is captured.
	ev := handler.expectNextEvent(t)
	_, ok := ev.(*event.DealingEvent)
	assert.Assert(t, ok)

	// No second delivery should follow (one goroutine, one event).
	select {
	case extra := <-handler.ch:
		t.Fatalf("unexpected duplicate event delivery: %T", extra)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestHandleKeyperSetAddedSubscribesFromEventBlockNumber(t *testing.T) {
	ksAddr := common.HexToAddress("0xaa")
	dkgAddr := common.HexToAddress("0xbb")
	backend := newFakeDKGBackend(dkgAddr)
	resolver := stubResolver(map[common.Address]common.Address{ksAddr: dkgAddr})

	handler := newRecordingHandler()
	s := &DKGEventSyncer{
		Log:                &logger.NoopLogger{},
		Handler:            handler.Handle,
		resolveDKGContract: resolver,
		bindDKGContract: func(addr common.Address) (dkgContractBackend, error) {
			if addr != dkgAddr {
				return nil, errors.Errorf("unknown DKG contract %s", addr.Hex())
			}
			return backend, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner()
	t.Cleanup(func() { cancel(); runner.Wait() })
	s.runner = runner

	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ksAddr,
		Eon:               2,
	}
	ev.Raw.BlockNumber = 1234

	s.handleKeyperSetAdded(ctx, ev)

	assert.DeepEqual(t, []common.Address{dkgAddr}, s.trackedDKGContractList())
	assert.Equal(t, uint64(1234), backend.dealingStart,
		"runtime subscription must start from the triggering KeyperSetAdded block number")

	// And the spawned goroutine should be live: delivering an event from the
	// fake backend must reach the handler.
	backend.emitDealing(t, &contract.DKGContractDealingSubmitted{KeyperSetIndex: 2, KeyperIndex: 0})
	got := handler.expectNextEvent(t)
	_, ok := got.(*event.DealingEvent)
	assert.Assert(t, ok, "expected *DealingEvent from runtime-spawned subscription, got %T", got)
}

func TestHandleKeyperSetAddedSkipsSubscriptionForExistingAddress(t *testing.T) {
	ksAddr := common.HexToAddress("0xaa")
	ksAddr2 := common.HexToAddress("0xab")
	dkgAddr := common.HexToAddress("0xbb")
	backend := newFakeDKGBackend(dkgAddr)
	resolver := stubResolver(map[common.Address]common.Address{ksAddr: dkgAddr, ksAddr2: dkgAddr})

	handler := newRecordingHandler()
	s := &DKGEventSyncer{
		Log:                &logger.NoopLogger{},
		Handler:            handler.Handle,
		resolveDKGContract: resolver,
		bindDKGContract: func(addr common.Address) (dkgContractBackend, error) {
			return backend, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner()
	t.Cleanup(func() { cancel(); runner.Wait() })
	s.runner = runner

	first := &bindings.KeyperSetManagerKeyperSetAdded{KeyperSetContract: ksAddr, Eon: 0}
	first.Raw.BlockNumber = 100
	s.handleKeyperSetAdded(ctx, first)
	firstStart := backend.dealingStart
	assert.Equal(t, uint64(100), firstStart)

	// A second keyper set added at a later block but pointing at the same
	// DKG contract must NOT re-subscribe and must not overwrite the
	// existing subscription's start block in the fake backend.
	second := &bindings.KeyperSetManagerKeyperSetAdded{KeyperSetContract: ksAddr2, Eon: 1}
	second.Raw.BlockNumber = 200
	s.handleKeyperSetAdded(ctx, second)

	// Tracked set still has one entry; backend.dealingStart unchanged.
	assert.DeepEqual(t, []common.Address{dkgAddr}, s.trackedDKGContractList())
	assert.Equal(t, firstStart, backend.dealingStart,
		"second KeyperSetAdded for an already-tracked DKG contract must not re-subscribe")

	// And there should be only one live subscription goroutine -- emitting
	// once delivers once.
	backend.emitDealing(t, &contract.DKGContractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: 0})
	_ = handler.expectNextEvent(t)
	select {
	case extra := <-handler.ch:
		t.Fatalf("unexpected duplicate event delivery: %T", extra)
	case <-time.After(100 * time.Millisecond):
	}
}
