package syncer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	gethevent "github.com/ethereum/go-ethereum/event"
	"github.com/shutter-network/shop-contracts/bindings"
	"gotest.tools/assert"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/logger"
)

// The fakes must satisfy the seam interfaces the EonPubKeySyncer talks to.
var (
	_ keyperSetCounter = (*fakeKeyperSetCounter)(nil)
	_ eonKeyBroadcast  = (*fakeEonKeyBroadcast)(nil)
)

// fakeKeyperSetCounter reports a fixed keyper-set count, which bounds the
// cold-start enumeration.
type fakeKeyperSetCounter struct{ n uint64 }

func (c *fakeKeyperSetCounter) GetNumKeyperSets(_ *bind.CallOpts) (uint64, error) {
	return c.n, nil
}

// fakeEonKeyBroadcast is an in-memory eonKeyBroadcast. keys holds the published
// key per keyper-set index; an index absent from the map models an unpublished
// key — GetEonKey returns empty bytes, the "not published yet" sentinel. The
// watch sink is captured on the first WatchEonKeyBroadcast call so tests can
// fire events.
type fakeEonKeyBroadcast struct {
	keys map[uint64][]byte

	mu         sync.Mutex
	sink       chan<- *bindings.KeyBroadcastContractEonKeyBroadcast
	watchStart uint64
}

func newFakeEonKeyBroadcast() *fakeEonKeyBroadcast {
	return &fakeEonKeyBroadcast{keys: map[uint64][]byte{}}
}

// publish records a published eon key at the given keyper-set index so the
// cold-start enumeration delivers it.
func (b *fakeEonKeyBroadcast) publish(index uint64, key []byte) {
	b.keys[index] = key
}

func (b *fakeEonKeyBroadcast) GetEonKey(_ *bind.CallOpts, eon uint64) ([]byte, error) {
	return b.keys[eon], nil
}

func (b *fakeEonKeyBroadcast) WatchEonKeyBroadcast(
	opts *bind.WatchOpts, sink chan<- *bindings.KeyBroadcastContractEonKeyBroadcast,
) (gethevent.Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.sink = sink
	if opts != nil && opts.Start != nil {
		b.watchStart = *opts.Start
	}
	return newFakeSubscription(), nil
}

// emitBroadcast fires an EonKeyBroadcast event on the captured watch sink.
func (b *fakeEonKeyBroadcast) emitBroadcast(t *testing.T, ev *bindings.KeyBroadcastContractEonKeyBroadcast) {
	t.Helper()
	b.mu.Lock()
	sink := b.sink
	b.mu.Unlock()
	assert.Assert(t, sink != nil, "eon key watch not set up yet")
	select {
	case sink <- ev:
	case <-time.After(time.Second):
		t.Fatal("eon key channel blocked")
	}
}

// recordingEonPubKeyHandler collects delivered eon keys on a buffered channel so
// tests can pull them off in order.
type recordingEonPubKeyHandler struct {
	ch chan *event.EonPublicKey
}

func newRecordingEonPubKeyHandler() *recordingEonPubKeyHandler {
	return &recordingEonPubKeyHandler{ch: make(chan *event.EonPublicKey, 32)}
}

func (h *recordingEonPubKeyHandler) Handle(_ context.Context, k *event.EonPublicKey) error {
	h.ch <- k
	return nil
}

func (h *recordingEonPubKeyHandler) expectNext(t *testing.T) *event.EonPublicKey {
	t.Helper()
	select {
	case k := <-h.ch:
		return k
	case <-time.After(2 * time.Second):
		t.Fatal("expected an eon key, got nothing")
		return nil
	}
}

func (h *recordingEonPubKeyHandler) expectNone(t *testing.T) {
	t.Helper()
	select {
	case k := <-h.ch:
		t.Fatalf("expected no eon key, got one at index %d", k.Eon)
	case <-time.After(200 * time.Millisecond):
	}
}

func newEonPubKeySyncer(
	count *fakeKeyperSetCounter, broadcast *fakeEonKeyBroadcast, startBlock uint64,
) (*EonPubKeySyncer, *recordingEonPubKeyHandler) {
	handler := newRecordingEonPubKeyHandler()
	s := &EonPubKeySyncer{
		Log:          &logger.NoopLogger{},
		StartBlock:   blockNumber(startBlock),
		Handler:      handler.Handle,
		manager:      count,
		keyBroadcast: broadcast,
	}
	return s, handler
}

func TestEonPubKeySyncerEnumeratesPublishedKeys(t *testing.T) {
	broadcast := newFakeEonKeyBroadcast()
	broadcast.publish(0, []byte("key-0"))
	broadcast.publish(1, []byte("key-1"))
	broadcast.publish(2, []byte("key-2"))

	s, handler := newEonPubKeySyncer(&fakeKeyperSetCounter{n: 3}, broadcast, 5)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	// Every published key is delivered, addressed by index, in order.
	for i := uint64(0); i < 3; i++ {
		k := handler.expectNext(t)
		assert.Equal(t, i, k.Eon, "key %d must be delivered with its index", i)
		assert.Equal(t, string([]byte{'k', 'e', 'y', '-', byte('0' + i)}), string(k.Key))
	}
	handler.expectNone(t)
}

func TestEonPubKeySyncerSkipsUnpublished(t *testing.T) {
	// Indices 0 and 2 are published; index 1 has no key yet. The unpublished
	// index must be skipped rather than errored (the resolved FIXME).
	broadcast := newFakeEonKeyBroadcast()
	broadcast.publish(0, []byte("key-0"))
	broadcast.publish(2, []byte("key-2"))

	s, handler := newEonPubKeySyncer(&fakeKeyperSetCounter{n: 3}, broadcast, 5)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner), "an unpublished key must not abort cold start")

	first := handler.expectNext(t)
	assert.Equal(t, uint64(0), first.Eon)
	second := handler.expectNext(t)
	assert.Equal(t, uint64(2), second.Eon, "index 1 is skipped, so index 2 comes next")
	handler.expectNone(t)
}

func TestEonPubKeySyncerNoKeysStartsCleanThenWatch(t *testing.T) {
	// Keyper sets exist but no DKG has completed, so no key is published yet.
	// The syncer must start cleanly (all indices skipped) and pick up keys via
	// the watch.
	broadcast := newFakeEonKeyBroadcast()
	s, handler := newEonPubKeySyncer(&fakeKeyperSetCounter{n: 2}, broadcast, 5)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))
	handler.expectNone(t)

	// A key broadcast later arrives via the watch.
	broadcast.emitBroadcast(t, &bindings.KeyBroadcastContractEonKeyBroadcast{
		Eon: 0,
		Key: []byte("key-0"),
		Raw: types.Log{BlockNumber: 60},
	})

	k := handler.expectNext(t)
	assert.Equal(t, uint64(0), k.Eon)
	assert.Equal(t, "key-0", string(k.Key))
}

func TestEonPubKeySyncerDeliversLaterKeyViaWatch(t *testing.T) {
	broadcast := newFakeEonKeyBroadcast()
	broadcast.publish(0, []byte("key-0"))

	s, handler := newEonPubKeySyncer(&fakeKeyperSetCounter{n: 1}, broadcast, 5)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))

	initial := handler.expectNext(t)
	assert.Equal(t, uint64(0), initial.Eon, "initial poll delivers the published key at index 0")

	// A key for a new keyper set arrives after start; the watch carries it.
	broadcast.emitBroadcast(t, &bindings.KeyBroadcastContractEonKeyBroadcast{
		Eon: 1,
		Key: []byte("key-1"),
		Raw: types.Log{BlockNumber: 110},
	})

	live := handler.expectNext(t)
	assert.Equal(t, uint64(1), live.Eon, "watch delivers the key at its own index")
	assert.Equal(t, "key-1", string(live.Key))
}

func TestEonPubKeySyncerWatchStartIsRequestedBlock(t *testing.T) {
	broadcast := newFakeEonKeyBroadcast()
	s, _ := newEonPubKeySyncer(&fakeKeyperSetCounter{n: 0}, broadcast, 4242)

	ctx, cancel := context.WithCancel(context.Background())
	runner := newFakeRunner(ctx)
	t.Cleanup(func() { cancel(); runner.Wait() })

	assert.NilError(t, s.Start(ctx, runner))
	assert.Equal(t, uint64(4242), broadcast.watchStart, "EonKeyBroadcast watch must start from the requested block")
}

func TestEonPubKeySyncerStartRequiresHandler(t *testing.T) {
	broadcast := newFakeEonKeyBroadcast()
	s := &EonPubKeySyncer{
		Log:          &logger.NoopLogger{},
		StartBlock:   blockNumber(1),
		manager:      &fakeKeyperSetCounter{n: 0},
		keyBroadcast: broadcast,
	}
	runner := newFakeRunner(context.Background())
	t.Cleanup(runner.Wait)
	err := s.Start(context.Background(), runner)
	assert.Assert(t, err != nil, "Start without a handler must error")
}
