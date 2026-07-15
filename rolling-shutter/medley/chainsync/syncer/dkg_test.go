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
	"github.com/shutter-network/contracts/v2/bindings/dkgcontract"
	"github.com/shutter-network/shop-contracts/bindings"
	"gotest.tools/assert"

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

func newSyncer(t *testing.T, resolver dkgContractResolver) (*DKGSyncer, *recordingLogger) {
	t.Helper()
	rec := newRecordingLogger()
	s := &DKGSyncer{
		Log:                rec,
		resolveDKGContract: resolver,
	}
	return s, rec
}

func TestTryTrackAddsNewAddress(t *testing.T) {
	s, rec := newSyncer(t, nil)
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")

	added := s.tryTrack(addr, 0)
	assert.Equal(t, true, added, "first insertion should be reported as newly added")
	assert.Equal(t, 0, rec.warnCount(), "non-zero address must not produce a warning")
	assert.DeepEqual(t, []common.Address{addr}, s.trackedDKGContractList())
	assert.Equal(t, uint64(1), s.getNumKnownKeyperSets(), "count must advance to ksi+1")
}

func TestTryTrackDeduplicatesButStillAdvancesCount(t *testing.T) {
	s, _ := newSyncer(t, nil)
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")

	addedFirst := s.tryTrack(addr, 0)
	addedSecond := s.tryTrack(addr, 1)

	assert.Equal(t, true, addedFirst)
	assert.Equal(t, false, addedSecond, "second insertion of the same address must be reported as not newly added")
	assert.Equal(t, 1, len(s.trackedDKGContractList()), "tracked set must contain a single entry after dedup")
	assert.Equal(t, uint64(2), s.getNumKnownKeyperSets(),
		"count must advance past both keyper set indices even though the address is shared")
}

func TestTryTrackZeroAddressWarnsSkipsAndAdvancesCount(t *testing.T) {
	s, rec := newSyncer(t, nil)

	added := s.tryTrack(common.Address{}, 7)

	assert.Equal(t, false, added, "zero address must not be added")
	assert.Equal(t, 0, len(s.trackedDKGContractList()))
	assert.Equal(t, 1, rec.warnCount(), "zero address must produce exactly one warning")
	assert.Equal(t, uint64(8), s.getNumKnownKeyperSets(),
		"count must advance even when the DKG contract address is zero")
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

	// Eon 0 is the expected next index given numKnownKeyperSets starts at 0, so
	// this exercises the normal (no-gap) recording path.
	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ksAddr,
		Eon:               0,
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

	// Eon 0 keeps this on the no-gap path; the single warning asserted below is
	// the zero-address warning, not a gap warning.
	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ksAddr,
		Eon:               0,
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

	// Eon 0 keeps this on the no-gap path so the test isolates resolver-error
	// tolerance for the triggering event.
	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: common.HexToAddress("0xaa"),
		Eon:               0,
	}
	s.handleKeyperSetAdded(context.Background(), ev)

	assert.Equal(t, 0, len(s.trackedDKGContractList()),
		"a resolver error must not result in any DKG contract being tracked")
}

func TestHandleKeyperSetAddedBackfillsGap(t *testing.T) {
	ks0 := common.HexToAddress("0xa0")
	ks1 := common.HexToAddress("0xa1")
	ks2 := common.HexToAddress("0xa2")
	dkg0 := common.HexToAddress("0xb0")
	dkg1 := common.HexToAddress("0xb1")
	dkg2 := common.HexToAddress("0xb2")

	keyperSets := map[uint64]common.Address{0: ks0, 1: ks1, 2: ks2}
	resolver := stubResolver(map[common.Address]common.Address{ks0: dkg0, ks1: dkg1, ks2: dkg2})
	backend0 := newFakeDKGBackend(dkg0)
	backend1 := newFakeDKGBackend(dkg1)
	backend2 := newFakeDKGBackend(dkg2)
	backends := map[common.Address]*fakeDKGBackend{dkg0: backend0, dkg1: backend1, dkg2: backend2}

	rec := newRecordingLogger()
	handler := newRecordingHandler()
	s := &DKGSyncer{
		Log:                rec,
		Handler:            handler.Handle,
		resolveDKGContract: resolver,
		bindDKGContract: func(addr common.Address) (dkgContractBackend, error) {
			b, ok := backends[addr]
			if !ok {
				return nil, errors.Errorf("unknown DKG contract %s", addr.Hex())
			}
			return b, nil
		},
		getKeyperSetAddress: func(_ *bind.CallOpts, index uint64) (common.Address, error) {
			addr, ok := keyperSets[index]
			if !ok {
				return common.Address{}, errors.Errorf("unknown keyper set index %d", index)
			}
			return addr, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })
	s.runner = runner
	s.startEventConsumer(ctx, runner)

	// numKnownKeyperSets starts at 0; a KeyperSetAdded event for index 2 implies
	// indices 0 and 1 were missed and must be backfilled before processing 2.
	ev := &bindings.KeyperSetManagerKeyperSetAdded{KeyperSetContract: ks2, Eon: 2}
	ev.Raw.BlockNumber = 500
	s.handleKeyperSetAdded(ctx, ev)

	// All three DKG contracts are tracked: the two backfilled plus the
	// triggering event's own contract.
	tracked := s.trackedDKGContractList()
	assert.Equal(t, 3, len(tracked))
	want := map[common.Address]struct{}{dkg0: {}, dkg1: {}, dkg2: {}}
	for _, addr := range tracked {
		_, ok := want[addr]
		assert.Assert(t, ok, "unexpected tracked contract %s", addr.Hex())
	}
	assert.Equal(t, uint64(3), s.getNumKnownKeyperSets(),
		"count must advance past the triggering keyper set index")

	// Backfilled syncers and the triggering syncer all start from the event block.
	assert.Equal(t, uint64(500), backend0.dealingStart, "backfilled syncer 0 must start at the triggering event block")
	assert.Equal(t, uint64(500), backend1.dealingStart, "backfilled syncer 1 must start at the triggering event block")
	assert.Equal(t, uint64(500), backend2.dealingStart, "triggering event syncer must start at its block")

	// Exactly one warning identifies the gap.
	assert.Equal(t, 1, rec.warnCount(), "exactly one gap warning expected")

	// Each spawned syncer is live: an event from each backend reaches the handler.
	backend0.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 0})
	backend1.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 1})
	backend2.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 2})
	got := map[uint64]bool{}
	for i := 0; i < 3; i++ {
		de, ok := handler.expectNextEvent(t).(*event.DealingEvent)
		assert.Assert(t, ok, "expected *DealingEvent")
		got[de.KeyperSetIndex] = true
	}
	assert.Assert(t, got[0] && got[1] && got[2], "events from all three syncers must be delivered")
}

func TestHandleKeyperSetAddedAlreadySeenIndexLogsWarningAndSkips(t *testing.T) {
	ksAddr := common.HexToAddress("0xaa")
	dkgAddr := common.HexToAddress("0xbb")
	backend := newFakeDKGBackend(dkgAddr)
	resolver := stubResolver(map[common.Address]common.Address{ksAddr: dkgAddr})

	rec := newRecordingLogger()
	handler := newRecordingHandler()
	s := &DKGSyncer{
		Log:                rec,
		Handler:            handler.Handle,
		resolveDKGContract: resolver,
		bindDKGContract: func(common.Address) (dkgContractBackend, error) {
			return backend, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })
	s.runner = runner
	s.startEventConsumer(ctx, runner)

	// Simulate three keyper sets already seen (non-zero addresses, no warnings).
	s.tryTrack(common.HexToAddress("0xc0"), 0)
	s.tryTrack(common.HexToAddress("0xc1"), 1)
	s.tryTrack(common.HexToAddress("0xc2"), 2)
	assert.Equal(t, uint64(3), s.getNumKnownKeyperSets())
	assert.Equal(t, 0, rec.warnCount())

	// An event for an already-seen index (1 < 3) is ignored with a warning.
	ev := &bindings.KeyperSetManagerKeyperSetAdded{KeyperSetContract: ksAddr, Eon: 1}
	ev.Raw.BlockNumber = 700
	s.handleKeyperSetAdded(ctx, ev)

	assert.Equal(t, 1, rec.warnCount(), "an already-seen index must log exactly one warning")
	assert.Equal(t, uint64(3), s.getNumKnownKeyperSets(), "count must be unchanged")
	assert.Equal(t, uint64(0), backend.dealingStart, "already-seen index must not start a syncer")
	for _, addr := range s.trackedDKGContractList() {
		assert.Assert(t, addr != dkgAddr, "ignored event must not track its DKG contract")
	}

	// No event is delivered for the ignored index.
	select {
	case ev := <-handler.ch:
		t.Fatalf("already-seen index must not deliver any event, got %T", ev)
	case <-time.After(100 * time.Millisecond):
	}
}

// gethLoggerInterface compile-time assertion: recordingLogger must satisfy the
// log.Logger interface so it can be wired into DKGSyncer.Log in real
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
	addr common.Address
	// succeededRetry may be nil; tests that do not care about the retry value
	// leave it unset and every success reports retry 0.
	succeeded      map[uint64]bool
	succeededRetry map[uint64]uint64

	// watchErrs maps an event-type key ("dealing", "accusation", "apology",
	// "successVote", "success") to an error the corresponding Watch* call should
	// return, letting tests exercise subscription-setup failure. A missing key
	// (the nil zero value) yields a working subscription.
	watchErrs map[string]error

	mu              sync.Mutex
	dealingSink     chan<- *dkgcontract.DkgcontractDealingSubmitted
	accusationSink  chan<- *dkgcontract.DkgcontractAccusationSubmitted
	apologySink     chan<- *dkgcontract.DkgcontractApologySubmitted
	successVoteSink chan<- *dkgcontract.DkgcontractSuccessVoteSubmitted
	successSink     chan<- *dkgcontract.DkgcontractDKGSucceeded

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

// SucceededAtRetry errors for keyper sets whose DKG did not succeed, matching
// the require(v != 0, "not succeeded") revert on chain.
func (b *fakeDKGBackend) SucceededAtRetry(_ *bind.CallOpts, ksi uint64) (uint64, error) {
	if !b.succeeded[ksi] {
		return 0, errors.New("not succeeded")
	}
	if b.succeededRetry == nil {
		return 0, nil
	}
	return b.succeededRetry[ksi], nil
}

func (b *fakeDKGBackend) WatchDealingSubmitted(
	opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractDealingSubmitted, _, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := b.watchErrs["dealing"]; err != nil {
		return nil, err
	}
	b.dealingSink = sink
	if opts != nil && opts.Start != nil {
		b.dealingStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) WatchAccusationSubmitted(
	opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractAccusationSubmitted, _, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := b.watchErrs["accusation"]; err != nil {
		return nil, err
	}
	b.accusationSink = sink
	if opts != nil && opts.Start != nil {
		b.accusationStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) WatchApologySubmitted(
	opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractApologySubmitted, _, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := b.watchErrs["apology"]; err != nil {
		return nil, err
	}
	b.apologySink = sink
	if opts != nil && opts.Start != nil {
		b.apologyStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) WatchSuccessVoteSubmitted(
	opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractSuccessVoteSubmitted, _, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := b.watchErrs["successVote"]; err != nil {
		return nil, err
	}
	b.successVoteSink = sink
	if opts != nil && opts.Start != nil {
		b.successVoteStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) WatchDKGSucceeded(
	opts *bind.WatchOpts, sink chan<- *dkgcontract.DkgcontractDKGSucceeded, _, _ []uint64,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := b.watchErrs["success"]; err != nil {
		return nil, err
	}
	b.successSink = sink
	if opts != nil && opts.Start != nil {
		b.successStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

func (b *fakeDKGBackend) emitDealing(t *testing.T, ev *dkgcontract.DkgcontractDealingSubmitted) {
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

func (b *fakeDKGBackend) emitSuccess(t *testing.T, ev *dkgcontract.DkgcontractDKGSucceeded) {
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

func (b *fakeDKGBackend) emitAccusation(t *testing.T, ev *dkgcontract.DkgcontractAccusationSubmitted) {
	t.Helper()
	b.mu.Lock()
	sink := b.accusationSink
	b.mu.Unlock()
	assert.Assert(t, sink != nil, "accusation subscription not set up yet")
	select {
	case sink <- ev:
	case <-time.After(time.Second):
		t.Fatal("accusation channel blocked")
	}
}

func (b *fakeDKGBackend) emitApology(t *testing.T, ev *dkgcontract.DkgcontractApologySubmitted) {
	t.Helper()
	b.mu.Lock()
	sink := b.apologySink
	b.mu.Unlock()
	assert.Assert(t, sink != nil, "apology subscription not set up yet")
	select {
	case sink <- ev:
	case <-time.After(time.Second):
		t.Fatal("apology channel blocked")
	}
}

func (b *fakeDKGBackend) emitSuccessVote(t *testing.T, ev *dkgcontract.DkgcontractSuccessVoteSubmitted) {
	t.Helper()
	b.mu.Lock()
	sink := b.successVoteSink
	b.mu.Unlock()
	assert.Assert(t, sink != nil, "success-vote subscription not set up yet")
	select {
	case sink <- ev:
	case <-time.After(time.Second):
		t.Fatal("success-vote channel blocked")
	}
}

// fakeRunner implements service.Runner for tests. Goroutines spawned via Go
// are tracked so the test can wait for them to drain after canceling its
// context. StartService starts each service synchronously under the runner's
// context, mirroring the real runner so that a DKGContractSyncer started by
// DKGSyncer comes up and observes context cancellation.
type fakeRunner struct {
	ctx context.Context
	wg  sync.WaitGroup
}

func newFakeRunner(ctx context.Context) *fakeRunner { return &fakeRunner{ctx: ctx} }

func (r *fakeRunner) Go(f func() error) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		_ = f()
	}()
}

func (r *fakeRunner) StartService(services ...service.Service) error {
	for _, s := range services {
		if err := s.Start(r.ctx, r); err != nil {
			return err
		}
	}
	return nil
}

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

func newSubscriptionSyncer(t *testing.T, backends map[common.Address]*fakeDKGBackend) (*DKGSyncer, *recordingHandler) {
	t.Helper()
	handler := newRecordingHandler()
	binder := func(addr common.Address) (dkgContractBackend, error) {
		b, ok := backends[addr]
		if !ok {
			return nil, errors.Errorf("unknown DKG contract %s", addr.Hex())
		}
		return b, nil
	}
	s := &DKGSyncer{
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
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	s.startEventConsumer(ctx, runner)
	assert.NilError(t, s.startContractSyncer(ctx, runner, addr, 42))

	backend.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{
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
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	s.startEventConsumer(ctx, runner)
	assert.NilError(t, s.startContractSyncer(ctx, runner, addr, 4242))

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
	// tryTrack(addr, 2) advances numKnownKeyperSets to 3 so initialSuccessesForContract
	// queries keyper set indices 0..2. Tracking addr itself is harmless here because
	// the test calls startContractSyncer directly rather than via the tracked list.
	s.tryTrack(addr, 2)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	s.startEventConsumer(ctx, runner)
	assert.NilError(t, s.startContractSyncer(ctx, runner, addr, 100))

	// Initial successes are enqueued before the live watcher starts, so with a
	// single FIFO consumer reading the handler channel first must yield the
	// SuccessEvents.
	ev1 := handler.expectNextEvent(t)
	_, ok := ev1.(*event.SuccessEvent)
	assert.Assert(t, ok, "first delivered event must be initial *SuccessEvent, got %T", ev1)

	ev2 := handler.expectNextEvent(t)
	_, ok = ev2.(*event.SuccessEvent)
	assert.Assert(t, ok, "second delivered event must be initial *SuccessEvent, got %T", ev2)

	// Only after both initial successes is the live event emitted; it must
	// arrive after them on the same channel.
	backend.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: 1})
	ev3 := handler.expectNextEvent(t)
	_, ok = ev3.(*event.DealingEvent)
	assert.Assert(t, ok, "live event after initial successes must be *DealingEvent, got %T", ev3)
}

// TestInitialSuccessesCarryRetryCounterFromSucceededAtRetry asserts that
// synthetic SuccessEvents emitted for already-completed DKGs at startup carry
// the RetryCounter the DKG actually succeeded at.
func TestInitialSuccessesCarryRetryCounterFromSucceededAtRetry(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	backend.succeeded[0] = true
	backend.succeeded[2] = true
	backend.succeededRetry = map[uint64]uint64{
		0: 0,
		2: 3,
	}

	s, handler := newSubscriptionSyncer(t, map[common.Address]*fakeDKGBackend{addr: backend})
	s.tryTrack(addr, 2)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	s.startEventConsumer(ctx, runner)
	assert.NilError(t, s.startContractSyncer(ctx, runner, addr, 100))

	got := map[uint64]uint64{}
	for i := 0; i < 2; i++ {
		ev := handler.expectNextEvent(t)
		se, ok := ev.(*event.SuccessEvent)
		assert.Assert(t, ok, "delivered event %d must be *SuccessEvent, got %T", i, ev)
		got[se.KeyperSetIndex] = se.RetryCounter
	}
	assert.Equal(t, uint64(0), got[0], "ksi 0 succeeded at retry 0")
	assert.Equal(t, uint64(3), got[2], "ksi 2 succeeded at retry 3")
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
	s.tryTrack(addr0, 0)
	s.tryTrack(addr1, 1)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	s.startEventConsumer(ctx, runner)
	for _, addr := range s.trackedDKGContractList() {
		assert.NilError(t, s.startContractSyncer(ctx, runner, addr, 100))
	}

	backend0.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: 1})
	backend1.emitSuccess(t, &dkgcontract.DkgcontractDKGSucceeded{KeyperSetIndex: 1, RetryCounter: 4})

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
	s.tryTrack(shared, 0)
	s.tryTrack(shared, 1)
	s.tryTrack(shared, 2)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	s.startEventConsumer(ctx, runner)
	// trackedDKGContractList must contain exactly one entry after dedup.
	tracked := s.trackedDKGContractList()
	assert.Equal(t, 1, len(tracked), "shared DKG contract must appear in tracked list exactly once")
	for _, addr := range tracked {
		assert.NilError(t, s.startContractSyncer(ctx, runner, addr, 100))
	}

	backend.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: 5})

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
	s := &DKGSyncer{
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
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })
	s.runner = runner
	s.startEventConsumer(ctx, runner)

	// Eon 0 is the expected next index (numKnownKeyperSets starts at 0), so this
	// exercises the normal no-gap path; only the subscription start block matters
	// here.
	ev := &bindings.KeyperSetManagerKeyperSetAdded{
		KeyperSetContract: ksAddr,
		Eon:               0,
	}
	ev.Raw.BlockNumber = 1234

	s.handleKeyperSetAdded(ctx, ev)

	assert.DeepEqual(t, []common.Address{dkgAddr}, s.trackedDKGContractList())
	assert.Equal(t, uint64(1234), backend.dealingStart,
		"runtime subscription must start from the triggering KeyperSetAdded block number")

	// And the spawned goroutine should be live: delivering an event from the
	// fake backend must reach the handler.
	backend.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 2, KeyperIndex: 0})
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
	s := &DKGSyncer{
		Log:                &logger.NoopLogger{},
		Handler:            handler.Handle,
		resolveDKGContract: resolver,
		bindDKGContract: func(_ common.Address) (dkgContractBackend, error) {
			return backend, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })
	s.runner = runner
	s.startEventConsumer(ctx, runner)

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
	// The count still advances past the second keyper set index even though no
	// new syncer was started for the shared DKG contract.
	assert.Equal(t, uint64(2), s.getNumKnownKeyperSets(),
		"numKnownKeyperSets must advance past the second keyper set index")

	// And there should be only one live subscription goroutine -- emitting
	// once delivers once.
	backend.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: 0})
	_ = handler.expectNextEvent(t)
	select {
	case extra := <-handler.ch:
		t.Fatalf("unexpected duplicate event delivery: %T", extra)
	case <-time.After(100 * time.Millisecond):
	}
}

// serialisationProbe is a DKGEventHandler that detects whether the syncer ever
// invokes it concurrently. On entry it increments an active-call counter and
// records the maximum ever observed; the brief sleep widens the window during
// which an overlapping call would be visible. With the fan-in channel funneling
// every event through a single consumer, maxActive must stay at 1.
type serialisationProbe struct {
	mu        sync.Mutex
	active    int
	maxActive int
	delivered int
	want      int
	done      chan struct{}
}

func newSerialisationProbe(want int) *serialisationProbe {
	return &serialisationProbe{want: want, done: make(chan struct{})}
}

func (p *serialisationProbe) Handle(_ context.Context, _ event.DKGEvent) error {
	p.mu.Lock()
	p.active++
	if p.active > p.maxActive {
		p.maxActive = p.active
	}
	p.mu.Unlock()

	// Widen the window during which a concurrent invocation would be observed.
	time.Sleep(2 * time.Millisecond)

	p.mu.Lock()
	p.active--
	p.delivered++
	if p.delivered == p.want {
		close(p.done)
	}
	p.mu.Unlock()
	return nil
}

// TestEventChannelSerialisesConcurrentHandlerCalls verifies that events produced
// concurrently by two independent DKGContractSyncer backends reach the real
// Handler one at a time: the fan-in channel and its single consumer goroutine
// must serialize all delivery so handler state needs no internal locking.
func TestEventChannelSerialisesConcurrentHandlerCalls(t *testing.T) {
	addr0 := common.HexToAddress("0xa0")
	addr1 := common.HexToAddress("0xa1")
	backend0 := newFakeDKGBackend(addr0)
	backend1 := newFakeDKGBackend(addr1)

	const perBackend = 5
	probe := newSerialisationProbe(2 * perBackend)
	s := &DKGSyncer{
		Log:     &logger.NoopLogger{},
		Handler: probe.Handle,
		bindDKGContract: func(addr common.Address) (dkgContractBackend, error) {
			switch addr {
			case addr0:
				return backend0, nil
			case addr1:
				return backend1, nil
			}
			return nil, errors.Errorf("unknown DKG contract %s", addr.Hex())
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	s.startEventConsumer(ctx, runner)
	assert.NilError(t, s.startContractSyncer(ctx, runner, addr0, 1))
	assert.NilError(t, s.startContractSyncer(ctx, runner, addr1, 1))

	// Pre-fill both backends' buffered subscription channels so their two
	// DKGContractSyncer goroutines drain and forward events concurrently. Without
	// the fan-in channel they would call the handler at the same time.
	for i := 0; i < perBackend; i++ {
		//nolint:gosec // G115: loop counter is bounded by perBackend
		backend0.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: uint64(i)})
		//nolint:gosec // G115: loop counter is bounded by perBackend
		backend1.emitSuccess(t, &dkgcontract.DkgcontractDKGSucceeded{KeyperSetIndex: 1, RetryCounter: uint64(i)})
	}

	select {
	case <-probe.done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not receive all events")
	}

	probe.mu.Lock()
	defer probe.mu.Unlock()
	assert.Equal(t, 2*perBackend, probe.delivered, "every emitted event must reach the handler")
	assert.Equal(t, 1, probe.maxActive, "handler must never be invoked concurrently")
}

// TestEventChannelInitialSuccessesPrecedeLiveEvents verifies the ordering
// guarantee the fan-in channel provides: a live event emitted while the initial
// successes for the same contract are still queued is still handled after them.
// The live event is emitted before any event is read from the handler, so only
// the FIFO consumer -- not synchronous delivery -- can be ordering them.
func TestEventChannelInitialSuccessesPrecedeLiveEvents(t *testing.T) {
	addr := common.HexToAddress("0xaa")
	backend := newFakeDKGBackend(addr)
	backend.succeeded[0] = true
	backend.succeeded[1] = true

	handler := newRecordingHandler()
	s := &DKGSyncer{
		Log:     &logger.NoopLogger{},
		Handler: handler.Handle,
		bindDKGContract: func(a common.Address) (dkgContractBackend, error) {
			if a != addr {
				return nil, errors.Errorf("unknown DKG contract %s", a.Hex())
			}
			return backend, nil
		},
	}
	// Two keyper set indices have succeeded, so initialSuccessesForContract
	// synthesizes two SuccessEvents to enqueue ahead of the live watcher.
	s.tryTrack(addr, 1)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	s.startEventConsumer(ctx, runner)
	assert.NilError(t, s.startContractSyncer(ctx, runner, addr, 100))

	// Emit a live event before reading anything: the initial successes were
	// enqueued during startContractSyncer (before the live watcher existed), so
	// FIFO ordering on the single channel must still hand them over first.
	backend.emitDealing(t, &dkgcontract.DkgcontractDealingSubmitted{KeyperSetIndex: 0, KeyperIndex: 9})

	first := handler.expectNextEvent(t)
	_, ok := first.(*event.SuccessEvent)
	assert.Assert(t, ok, "first delivered event must be an initial *SuccessEvent, got %T", first)

	second := handler.expectNextEvent(t)
	_, ok = second.(*event.SuccessEvent)
	assert.Assert(t, ok, "second delivered event must be an initial *SuccessEvent, got %T", second)

	third := handler.expectNextEvent(t)
	_, ok = third.(*event.DealingEvent)
	assert.Assert(t, ok, "the live event must be delivered after the initial successes, got %T", third)
}
