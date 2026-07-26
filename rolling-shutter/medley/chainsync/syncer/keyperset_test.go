package syncer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	gethevent "github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/logger"
)

// The fakes must satisfy the seam interfaces the KeyperSetSyncer talks to.
var (
	_ keyperSetManager = (*fakeKeyperSetManager)(nil)
	_ keyperSetReader  = fakeKeyperSetReader{}
)

// fakeKeyperSet is the canned per-index state a fake manager hands out: the
// KeyperSet contract address, its activation block, and the per-set reads.
type fakeKeyperSet struct {
	addr       common.Address
	activation uint64
	finalized  bool
	members    []common.Address
	threshold  uint64
}

// fakeKeyperSetManager is an in-memory keyperSetManager. byIndex backs the
// cold-start enumeration; readers backs the per-set-read binder (and holds
// watch-only sets that are not part of the initial enumeration). The watch sink
// is captured on the first WatchKeyperSetAdded call so tests can fire events.
type fakeKeyperSetManager struct {
	byIndex []fakeKeyperSet
	readers map[common.Address]fakeKeyperSet

	mu         sync.Mutex
	sink       chan<- *bindings.KeyperSetManagerKeyperSetAdded
	watchStart uint64
}

func newFakeKeyperSetManager() *fakeKeyperSetManager {
	return &fakeKeyperSetManager{readers: map[common.Address]fakeKeyperSet{}}
}

// addInitial registers a keyper set that the cold-start poll enumerates, at the
// next free index.
func (m *fakeKeyperSetManager) addInitial(ks fakeKeyperSet) {
	m.byIndex = append(m.byIndex, ks)
	m.readers[ks.addr] = ks
}

// register makes a keyper set resolvable by the binder without adding it to the
// cold-start enumeration — used for sets that only arrive via the watch.
func (m *fakeKeyperSetManager) register(ks fakeKeyperSet) {
	m.readers[ks.addr] = ks
}

func (m *fakeKeyperSetManager) GetNumKeyperSets(_ *bind.CallOpts) (uint64, error) {
	return uint64(len(m.byIndex)), nil
}

func (m *fakeKeyperSetManager) GetKeyperSetAddress(_ *bind.CallOpts, index uint64) (common.Address, error) {
	if index >= uint64(len(m.byIndex)) {
		return common.Address{}, errors.Errorf("keyper set index %d out of range", index)
	}
	return m.byIndex[index].addr, nil
}

func (m *fakeKeyperSetManager) GetKeyperSetActivationBlock(_ *bind.CallOpts, index uint64) (uint64, error) {
	if index >= uint64(len(m.byIndex)) {
		return 0, errors.Errorf("keyper set index %d out of range", index)
	}
	return m.byIndex[index].activation, nil
}

func (m *fakeKeyperSetManager) WatchKeyperSetAdded(
	opts *bind.WatchOpts, sink chan<- *bindings.KeyperSetManagerKeyperSetAdded,
) (gethevent.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sink = sink
	if opts != nil && opts.Start != nil {
		m.watchStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

// binder returns a keyperSetBinder resolving addresses against the registered
// readers.
func (m *fakeKeyperSetManager) binder() keyperSetBinder {
	return func(addr common.Address) (keyperSetReader, error) {
		ks, ok := m.readers[addr]
		if !ok {
			return nil, errors.Errorf("no keyper set at %s", addr.Hex())
		}
		return fakeKeyperSetReader{ks}, nil
	}
}

// emitAdded fires a KeyperSetAdded event on the captured watch sink.
func (m *fakeKeyperSetManager) emitAdded(t *testing.T, ev *bindings.KeyperSetManagerKeyperSetAdded) {
	t.Helper()
	m.mu.Lock()
	sink := m.sink
	m.mu.Unlock()
	assert.Assert(t, sink != nil, "keyper set watch not set up yet")
	select {
	case sink <- ev:
	case <-time.After(time.Second):
		t.Fatal("keyper set channel blocked")
	}
}

type fakeKeyperSetReader struct{ ks fakeKeyperSet }

func (r fakeKeyperSetReader) IsFinalized(_ *bind.CallOpts) (bool, error) { return r.ks.finalized, nil }
func (r fakeKeyperSetReader) GetMembers(_ *bind.CallOpts) ([]common.Address, error) {
	return r.ks.members, nil
}

func (r fakeKeyperSetReader) GetThreshold(_ *bind.CallOpts) (uint64, error) {
	return r.ks.threshold, nil
}

// recordingKeyperSetHandler collects delivered keyper sets on a buffered channel
// so tests can pull them off in order.
type recordingKeyperSetHandler struct {
	ch chan *event.KeyperSet
}

func newRecordingKeyperSetHandler() *recordingKeyperSetHandler {
	return &recordingKeyperSetHandler{ch: make(chan *event.KeyperSet, 32)}
}

func (h *recordingKeyperSetHandler) Handle(_ context.Context, ks *event.KeyperSet) error {
	h.ch <- ks
	return nil
}

func (h *recordingKeyperSetHandler) expectNext(t *testing.T) *event.KeyperSet {
	t.Helper()
	select {
	case ks := <-h.ch:
		return ks
	case <-time.After(2 * time.Second):
		t.Fatal("expected a keyper set, got nothing")
		return nil
	}
}

func (h *recordingKeyperSetHandler) expectNone(t *testing.T) {
	t.Helper()
	select {
	case ks := <-h.ch:
		t.Fatalf("expected no keyper set, got one with index %d", ks.Eon)
	case <-time.After(200 * time.Millisecond):
	}
}

func fakeSet(hexAddr string, activation, threshold uint64, members ...common.Address) fakeKeyperSet {
	return fakeKeyperSet{
		addr:       common.HexToAddress(hexAddr),
		activation: activation,
		finalized:  true,
		members:    members,
		threshold:  threshold,
	}
}

func newKeyperSetSyncer(mgr *fakeKeyperSetManager, startBlock uint64) (*KeyperSetSyncer, *recordingKeyperSetHandler) {
	handler := newRecordingKeyperSetHandler()
	s := &KeyperSetSyncer{
		Log:           &logger.NoopLogger{},
		StartBlock:    blockNumber(startBlock),
		Handler:       handler.Handle,
		manager:       mgr,
		bindKeyperSet: mgr.binder(),
	}
	return s, handler
}

func TestKeyperSetSyncerEnumeratesAllOnColdStart(t *testing.T) {
	mgr := newFakeKeyperSetManager()
	mgr.addInitial(fakeSet("0xa0", 10, 1, common.HexToAddress("0x01")))
	mgr.addInitial(fakeSet("0xa1", 20, 2, common.HexToAddress("0x02"), common.HexToAddress("0x03")))
	mgr.addInitial(fakeSet("0xa2", 30, 3))

	s, handler := newKeyperSetSyncer(mgr, 5)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	// Every registered set is delivered, addressed by index, in order.
	for i := uint64(0); i < 3; i++ {
		ks := handler.expectNext(t)
		assert.Equal(t, i, ks.Eon, "set %d must be delivered with its index", i)
	}
	handler.expectNone(t)
}

func TestKeyperSetSyncerEnumeratePreservesFields(t *testing.T) {
	mgr := newFakeKeyperSetManager()
	memberA := common.HexToAddress("0x02")
	memberB := common.HexToAddress("0x03")
	mgr.addInitial(fakeSet("0xa1", 20, 2, memberA, memberB))

	s, handler := newKeyperSetSyncer(mgr, 5)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	ks := handler.expectNext(t)
	assert.Equal(t, uint64(0), ks.Eon)
	assert.Equal(t, uint64(20), ks.ActivationBlock)
	assert.Equal(t, uint64(2), ks.Threshold)
	assert.Equal(t, common.HexToAddress("0xa1"), ks.Contract)
	assert.Equal(t, 2, len(ks.Members))
	assert.Equal(t, memberA, ks.Members[0])
	assert.Equal(t, memberB, ks.Members[1])
}

func TestKeyperSetSyncerEmptyContractStartsClean(t *testing.T) {
	mgr := newFakeKeyperSetManager()
	s, handler := newKeyperSetSyncer(mgr, 5)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	// An empty contract must not abort the cold start.
	assert.NilError(t, s.Start(ctx, runner))
	handler.expectNone(t)

	// A set registered later arrives via the watch.
	added := fakeSet("0xb0", 50, 1, common.HexToAddress("0x04"))
	mgr.register(added)
	mgr.emitAdded(t, &bindings.KeyperSetManagerKeyperSetAdded{
		ActivationBlock:   added.activation,
		KeyperSetContract: added.addr,
		Eon:               0,
		Raw:               types.Log{BlockNumber: 60},
	})

	ks := handler.expectNext(t)
	assert.Equal(t, uint64(0), ks.Eon)
	assert.Equal(t, added.addr, ks.Contract)
}

func TestKeyperSetSyncerDeliversLaterSetViaWatch(t *testing.T) {
	mgr := newFakeKeyperSetManager()
	mgr.addInitial(fakeSet("0xa0", 10, 1, common.HexToAddress("0x01")))

	s, handler := newKeyperSetSyncer(mgr, 5)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	initial := handler.expectNext(t)
	assert.Equal(t, uint64(0), initial.Eon, "initial poll delivers index 0")

	// A new set arrives with index 1 after start; the watch carries the index.
	added := fakeSet("0xa1", 100, 2, common.HexToAddress("0x02"))
	mgr.register(added)
	mgr.emitAdded(t, &bindings.KeyperSetManagerKeyperSetAdded{
		ActivationBlock:   added.activation,
		KeyperSetContract: added.addr,
		Eon:               1,
		Raw:               types.Log{BlockNumber: 110},
	})

	live := handler.expectNext(t)
	assert.Equal(t, uint64(1), live.Eon, "watch delivers the set at its own index")
	assert.Equal(t, added.addr, live.Contract)
	assert.Equal(t, uint64(100), live.ActivationBlock)
}

func TestKeyperSetSyncerDeliversNotYetActiveInitialSet(t *testing.T) {
	// The single registered set activates far in the future relative to the
	// start block. The old design queried "which set is active at start block?"
	// and aborted with NoActiveKeyperSet; the by-index poll delivers it anyway.
	mgr := newFakeKeyperSetManager()
	mgr.addInitial(fakeSet("0xa0", 100000, 1, common.HexToAddress("0x01")))

	s, handler := newKeyperSetSyncer(mgr, 5)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner), "cold start must not abort when no set is active yet")

	ks := handler.expectNext(t)
	assert.Equal(t, uint64(0), ks.Eon)
	assert.Equal(t, uint64(100000), ks.ActivationBlock)
}

func TestKeyperSetSyncerWatchStartIsRequestedBlock(t *testing.T) {
	mgr := newFakeKeyperSetManager()
	s, _ := newKeyperSetSyncer(mgr, 4242)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))
	assert.Equal(t, uint64(4242), mgr.watchStart, "KeyperSetAdded watch must start from the requested block")
}

func TestKeyperSetSyncerStartRequiresHandler(t *testing.T) {
	mgr := newFakeKeyperSetManager()
	s := &KeyperSetSyncer{
		Log:           &logger.NoopLogger{},
		StartBlock:    blockNumber(1),
		manager:       mgr,
		bindKeyperSet: mgr.binder(),
	}
	runner := newFakeRunner(context.Background())
	t.Cleanup(runner.Wait)
	err := s.Start(context.Background(), runner)
	assert.Assert(t, err != nil, "Start without a handler must error")
}
