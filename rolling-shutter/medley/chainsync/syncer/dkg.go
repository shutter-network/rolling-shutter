package syncer

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethevent "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	keypersetBindings "github.com/shutter-network/contracts/v2/bindings/keyperset"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

// keyperSetIndexer is the subset of *bindings.KeyperSetManager used by the
// DKG contract discovery scan. Pulling it out as an interface lets tests
// substitute a fake manager without spinning up a simulated chain.
type keyperSetIndexer interface {
	GetNumKeyperSets(opts *bind.CallOpts) (uint64, error)
	GetKeyperSetAddress(opts *bind.CallOpts, index uint64) (common.Address, error)
}

// dkgContractResolver resolves a KeyperSet contract address to the DKG contract
// address it points at (via KeyperSet.getDKGContract()).
type dkgContractResolver func(ctx context.Context, opts *bind.CallOpts, keyperSetAddr common.Address) (common.Address, error)

// dkgContractBackend is the subset of *contract.DKGContract used by the
// per-contract subscription goroutine: the Succeeded(ksi) read for initial
// success synthesis and the five bulletin-board event subscriptions. Pulling
// it out as an interface lets tests substitute a fake backend that emits
// canned events without touching a simulated chain.
type dkgContractBackend interface {
	Succeeded(opts *bind.CallOpts, keyperSetIndex uint64) (bool, error)
	WatchDealingSubmitted(opts *bind.WatchOpts, sink chan<- *contract.DKGContractDealingSubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (gethevent.Subscription, error)
	WatchAccusationSubmitted(opts *bind.WatchOpts, sink chan<- *contract.DKGContractAccusationSubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (gethevent.Subscription, error)
	WatchApologySubmitted(opts *bind.WatchOpts, sink chan<- *contract.DKGContractApologySubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (gethevent.Subscription, error)
	WatchSuccessVoteSubmitted(opts *bind.WatchOpts, sink chan<- *contract.DKGContractSuccessVoteSubmitted, keyperSetIndex []uint64, retryCounter []uint64, keyperIndex []uint64) (gethevent.Subscription, error)
	WatchDKGSucceeded(opts *bind.WatchOpts, sink chan<- *contract.DKGContractDKGSucceeded, keyperSetIndex []uint64, retryCounter []uint64) (gethevent.Subscription, error)
}

// dkgContractBinder constructs a dkgContractBackend for a given DKG contract
// address. The production implementation binds a *contract.DKGContract; tests
// substitute an in-memory fake.
type dkgContractBinder func(addr common.Address) (dkgContractBackend, error)

// DKGSyncer discovers DKG contract addresses autonomously by scanning
// every keyper set in KeyperSetManager and subscribing to KeyperSetAdded.
// For each unique non-zero DKG contract address it starts one DKGContractSyncer
// that watches the five bulletin-board event types (DealingSubmitted,
// AccusationSubmitted, ApologySubmitted, SuccessVoteSubmitted, DKGSucceeded) and
// forwards them to the shared Handler.
//
// Discovery and subscription are deduplicated by DKG contract address, so
// multiple keyper sets sharing one DKG contract result in exactly one
// DKGContractSyncer and no double event delivery.
type DKGSyncer struct {
	Client           client.Client
	KeyperSetManager *bindings.KeyperSetManager
	Log              log.Logger
	StartBlock       *number.BlockNumber
	Handler          event.DKGEventHandler

	// resolveDKGContract is the function used to read the DKG contract address
	// from a KeyperSet contract. It is overridable so tests can avoid binding
	// to a real contract; production callers leave it nil and Start() fills in
	// the default binding-backed implementation.
	resolveDKGContract dkgContractResolver

	// bindDKGContract constructs a dkgContractBackend for a DKG contract
	// address. Overridable for tests; Start() fills in the default
	// binding-backed implementation when nil.
	bindDKGContract dkgContractBinder

	// getKeyperSetAddress reads the KeyperSet contract address for a given
	// Keyper Set Index. It is the subset of KeyperSetManager the gap-backfill
	// path needs; overridable so tests can inject canned addresses. Start()
	// fills in s.KeyperSetManager.GetKeyperSetAddress when nil.
	getKeyperSetAddress func(opts *bind.CallOpts, index uint64) (common.Address, error)

	// runner is captured from Start() so that handleKeyperSetAdded can spawn
	// new per-contract subscription goroutines as new DKG contract addresses
	// are discovered at runtime.
	runner service.Runner

	keyperSetAddedCh chan *bindings.KeyperSetManagerKeyperSetAdded

	// eventCh is the fan-in channel through which every DKG event -- initial
	// successes synthesised at startup and live bulletin-board events from all
	// DKGContractSyncer goroutines -- is funnelled to a single consumer
	// goroutine. The consumer calls the real Handler sequentially, so handler
	// state needs no internal locking. It is created in Start() with the shared
	// channelSize buffer.
	eventCh chan event.DKGEvent

	trackedMu           sync.Mutex
	trackedDKGContracts map[common.Address]struct{}
	// numKnownKeyperSets is the number of Keyper Set Indices the node has
	// observed in KeyperSetManager: an exclusive upper bound used to iterate
	// the Succeeded(ksi) queries when a new DKG contract subscription is
	// started. It is monotonically non-decreasing because keyper set indices
	// in KeyperSetManager are append-only.
	numKnownKeyperSets uint64
}

func (s *DKGSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}

	if s.StartBlock.IsLatest() {
		latest, err := s.Client.BlockNumber(ctx)
		if err != nil {
			return err
		}
		s.StartBlock.SetUint64(latest)
	}

	if s.resolveDKGContract == nil {
		s.resolveDKGContract = defaultDKGContractResolver(s.Client)
	}
	if s.bindDKGContract == nil {
		s.bindDKGContract = defaultDKGContractBinder(s.Client)
	}
	if s.getKeyperSetAddress == nil {
		s.getKeyperSetAddress = s.KeyperSetManager.GetKeyperSetAddress
	}
	s.runner = runner
	s.startEventConsumer(ctx, runner)

	startBlock := *s.StartBlock.ToUInt64Ptr()
	callOpts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	if err := s.scanInitialDKGContracts(ctx, callOpts, s.KeyperSetManager); err != nil {
		return errors.Wrap(err, "initial DKG contract scan")
	}

	for _, addr := range s.trackedDKGContractList() {
		if err := s.startContractSyncer(ctx, runner, addr, startBlock); err != nil {
			return errors.Wrapf(err, "start syncer for DKG contract %s", addr.Hex())
		}
	}

	watchOpts := &bind.WatchOpts{
		Start:   &startBlock,
		Context: ctx,
	}
	s.keyperSetAddedCh = make(chan *bindings.KeyperSetManagerKeyperSetAdded, channelSize)
	keyperAddedSub, err := s.KeyperSetManager.WatchKeyperSetAdded(watchOpts, s.keyperSetAddedCh)
	if err != nil {
		return errors.Wrap(err, "watch KeyperSetAdded")
	}

	runner.Go(func() error {
		err := s.watchKeyperSetAdded(ctx, keyperAddedSub.Err())
		if err != nil {
			s.Log.Error("error watching KeyperSetAdded events", "error", err.Error())
		}
		keyperAddedSub.Unsubscribe()
		return err
	})
	return nil
}

// startEventConsumer initialises the fan-in channel and starts the single
// consumer goroutine that drains it, calling the real Handler sequentially.
// Funnelling every DKG event through one consumer guarantees the Handler is
// never invoked concurrently, so DKGEventHandler implementations need no
// internal locking. Handler errors are logged and draining continues, matching
// the deliver() behaviour DKGContractSyncer used before this fan-in existed.
// The consumer exits on context cancellation; producers send via enqueueEvent,
// whose send is context-aware, so shutdown never deadlocks.
func (s *DKGSyncer) startEventConsumer(ctx context.Context, runner service.Runner) {
	s.eventCh = make(chan event.DKGEvent, channelSize)
	runner.Go(func() error {
		for {
			select {
			case ev := <-s.eventCh:
				if err := s.Handler(ctx, ev); err != nil {
					s.Log.Error("handler for DKG event errored", "error", err.Error())
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
}

// enqueueEvent hands a DKG event to the fan-in channel for serialised delivery
// by the consumer goroutine. It is the Handler passed to every DKGContractSyncer
// and the delivery path for initial-success events. The send is context-aware so
// that a producer never blocks forever when the node is shutting down and the
// consumer has already exited.
func (s *DKGSyncer) enqueueEvent(ctx context.Context, ev event.DKGEvent) error {
	select {
	case s.eventCh <- ev:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// scanInitialDKGContracts iterates every keyper set known to the manager and
// records the DKG contract address each one points at via tryTrack. Failures to
// read an individual keyper set's DKG contract (RPC error) are surfaced; zero
// addresses are logged and skipped inside tryTrack, mirroring the runtime
// behaviour for KeyperSetAdded. As a side effect, numKnownKeyperSets advances to
// the number of keyper sets visited. The tryTrack return value is ignored here;
// Start() iterates trackedDKGContractList() afterwards to spawn syncers.
func (s *DKGSyncer) scanInitialDKGContracts(
	ctx context.Context,
	opts *bind.CallOpts,
	indexer keyperSetIndexer,
) error {
	if err := guardCallOpts(opts, false); err != nil {
		return err
	}
	numKS, err := indexer.GetNumKeyperSets(opts)
	if err != nil {
		return errors.Wrap(err, "get num keyper sets")
	}
	for i := uint64(0); i < numKS; i++ {
		ksAddr, err := indexer.GetKeyperSetAddress(opts, i)
		if err != nil {
			return errors.Wrapf(err, "get keyper set address %d", i)
		}
		dkgAddr, err := s.resolveDKGContract(ctx, opts, ksAddr)
		if err != nil {
			return errors.Wrapf(err, "resolve DKG contract for keyper set %d (%s)", i, ksAddr.Hex())
		}
		s.tryTrack(dkgAddr, i)
	}
	return nil
}

// tryTrack is the single write path for both trackedDKGContracts and
// numKnownKeyperSets, performing all mutation under one trackedMu acquisition so
// that no observer ever sees the count advanced without the contract recorded or
// vice versa. It always advances numKnownKeyperSets to max(current, ksi+1).
//
// It returns true only when addr is a new, non-zero DKG contract address (i.e. a
// DKGContractSyncer should be started for it). A zero address logs a warning and
// returns false; an already-tracked address returns false. In both of those
// cases the count still advances, so shared-contract or unconfigured keyper sets
// never leave a gap in numKnownKeyperSets.
func (s *DKGSyncer) tryTrack(addr common.Address, ksi uint64) bool {
	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()

	if ksi+1 > s.numKnownKeyperSets {
		s.numKnownKeyperSets = ksi + 1
	}

	if (addr == common.Address{}) {
		s.Log.Warn(
			"keyper set has no DKG contract configured; skipping",
			"keyper-set-index", ksi,
		)
		return false
	}

	if s.trackedDKGContracts == nil {
		s.trackedDKGContracts = map[common.Address]struct{}{}
	}
	if _, exists := s.trackedDKGContracts[addr]; exists {
		return false
	}
	s.trackedDKGContracts[addr] = struct{}{}
	return true
}

// trackedDKGContractList returns a snapshot of the currently tracked DKG
// contract addresses. Intended for tests and for Start() to iterate the set
// when spawning per-contract subscription goroutines.
func (s *DKGSyncer) trackedDKGContractList() []common.Address {
	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()
	out := make([]common.Address, 0, len(s.trackedDKGContracts))
	for addr := range s.trackedDKGContracts {
		out = append(out, addr)
	}
	return out
}

func (s *DKGSyncer) getNumKnownKeyperSets() uint64 {
	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()
	return s.numKnownKeyperSets
}

// startContractSyncer delivers the initial success state for the given DKG
// contract and then starts a DKGContractSyncer to watch it for live events.
// It first queries Succeeded(ksi) for every known keyper set index at
// startBlock and enqueues synthetic SuccessEvents for already-completed
// instances onto the fan-in channel. It then constructs a DKGContractSyncer
// bound to the same backend -- whose Handler is the same fan-in send -- and
// starts it via runner.StartService, which sets up the five bulletin-board
// subscriptions from startBlock and forwards live events onto the fan-in
// channel too. Because initial successes are enqueued before the live watcher
// starts and a single consumer drains the channel in FIFO order, catch-up
// state always reaches the real Handler before live events for the contract.
//
// startBlock is the height from which live events are watched. At startup
// it is the resolved StartBlock; at runtime it is the block of the triggering
// KeyperSetAdded event so events fired between keyper set registration and
// subscription startup are not missed.
func (s *DKGSyncer) startContractSyncer(
	ctx context.Context,
	runner service.Runner,
	addr common.Address,
	startBlock uint64,
) error {
	backend, err := s.bindDKGContract(addr)
	if err != nil {
		return errors.Wrapf(err, "bind DKG contract %s", addr.Hex())
	}

	initial, err := s.initialSuccessesForContract(ctx, backend, startBlock)
	if err != nil {
		return errors.Wrapf(err, "initial successes for DKG contract %s", addr.Hex())
	}
	// Enqueue initial successes before starting the live watcher so that, with a
	// single FIFO consumer, catch-up state for this contract is always handled
	// before any live event from the same contract.
	for _, ev := range initial {
		if err := s.enqueueEvent(ctx, ev); err != nil {
			return errors.Wrapf(err, "enqueue initial success for DKG contract %s", addr.Hex())
		}
	}

	contractSyncer := &DKGContractSyncer{
		Addr:       addr,
		Log:        s.Log,
		Handler:    s.enqueueEvent,
		StartBlock: number.NewBlockNumber(&startBlock),
		backend:    backend,
	}
	if err := runner.StartService(contractSyncer); err != nil {
		return errors.Wrapf(err, "start DKG contract syncer %s", addr.Hex())
	}
	return nil
}

// initialSuccessesForContract queries DKGContract.succeeded(ksi) on the given
// contract for each known keyper set index. Already-succeeded instances are
// returned as synthetic SuccessEvents so the local cache (e.g. dkg_result)
// can be populated before any live events for the same contract arrive.
func (s *DKGSyncer) initialSuccessesForContract(
	ctx context.Context,
	backend dkgContractBackend,
	startBlock uint64,
) ([]event.DKGEvent, error) {
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(startBlock),
	}
	if err := guardCallOpts(opts, false); err != nil {
		return nil, err
	}
	numKS := s.getNumKnownKeyperSets()
	events := make([]event.DKGEvent, 0, numKS)
	for i := uint64(0); i < numKS; i++ {
		succeeded, err := backend.Succeeded(opts, i)
		if err != nil {
			return nil, errors.Wrapf(err, "query succeeded for keyper set %d", i)
		}
		if !succeeded {
			continue
		}
		events = append(events, &event.SuccessEvent{
			KeyperSetIndex: i,
			AtBlockNumber:  number.BigToBlockNumber(opts.BlockNumber),
		})
	}
	return events, nil
}

// watchKeyperSetAdded drains the KeyperSetAdded subscription channel and
// dispatches each event to handleKeyperSetAdded for DKG contract discovery
// and subscription spawning. RPC failures and zero addresses are tolerated;
// a single bad keyper set must not bring down DKG event syncing for the rest.
func (s *DKGSyncer) watchKeyperSetAdded(ctx context.Context, subErr <-chan error) error {
	for {
		select {
		case ev, ok := <-s.keyperSetAddedCh:
			if !ok {
				return nil
			}
			s.handleKeyperSetAdded(ctx, ev)
		case err := <-subErr:
			if err != nil {
				return errors.Wrap(err, "KeyperSetAdded subscription")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handleKeyperSetAdded processes a KeyperSetAdded event for DKG contract
// discovery, detecting and repairing gaps in the event stream first.
//
// The incoming Keyper Set Index (ev.Eon) is compared against numKnownKeyperSets:
//
//   - smaller (already seen): the event is a duplicate or arrived out of order;
//     a warning is logged and the event is ignored.
//   - larger (gap): one or more KeyperSetAdded events were missed. A warning is
//     logged and every missed index in [numKnownKeyperSets, ev.Eon) is
//     backfilled -- read, resolved, tracked, and (if newly added) synced from
//     the triggering event's block -- before the current event is processed.
//   - equal (expected): the event is processed directly.
//
// Backfill runs synchronously in the watchKeyperSetAdded goroutine; gaps are
// expected to be rare, so blocking the event loop for its duration is accepted.
//
// For the triggering event itself, the DKG contract address is resolved,
// recorded via tryTrack, and -- if newly added and a runner is available -- a
// per-contract subscription goroutine is spawned from the block of the event.
// RPC failures and zero addresses are logged but do not abort the event loop.
// On a resolve error the state (including numKnownKeyperSets) is left untouched,
// trading partial-update semantics for a single write path through tryTrack.
func (s *DKGSyncer) handleKeyperSetAdded(ctx context.Context, ev *bindings.KeyperSetManagerKeyperSetAdded) {
	opts := logToCallOpts(ctx, &ev.Raw)

	switch known := s.getNumKnownKeyperSets(); {
	case ev.Eon < known:
		s.Log.Warn(
			"received KeyperSetAdded for an already-seen keyper set index; ignoring",
			"expected-keyper-set-index", known,
			"received-keyper-set-index", ev.Eon,
		)
		return
	case ev.Eon > known:
		s.Log.Warn(
			"gap in KeyperSetAdded event stream; backfilling missed keyper set indices",
			"expected-keyper-set-index", known,
			"received-keyper-set-index", ev.Eon,
		)
		for i := known; i < ev.Eon; i++ {
			s.backfillKeyperSet(ctx, opts, i, ev.Raw.BlockNumber)
		}
	}

	dkgAddr, err := s.resolveDKGContract(ctx, opts, ev.KeyperSetContract)
	if err != nil {
		s.Log.Error(
			"could not resolve DKG contract for new keyper set",
			"error", err.Error(),
			"keyper-set", ev.KeyperSetContract.Hex(),
			"eon", ev.Eon,
		)
		return
	}
	if !s.tryTrack(dkgAddr, ev.Eon) {
		return
	}
	if s.runner == nil || s.bindDKGContract == nil {
		return
	}
	if err := s.startContractSyncer(ctx, s.runner, dkgAddr, ev.Raw.BlockNumber); err != nil {
		s.Log.Error(
			"could not start DKG contract subscription for new keyper set",
			"error", err.Error(),
			"dkg-contract", dkgAddr.Hex(),
			"keyper-set", ev.KeyperSetContract.Hex(),
			"eon", ev.Eon,
		)
	}
}

// backfillKeyperSet recovers a single Keyper Set Index that was missing from the
// KeyperSetAdded event stream: it reads the KeyperSet contract address, resolves
// its DKG contract, records it via tryTrack, and -- if the address is newly
// added and a runner is available -- starts a DKGContractSyncer from startBlock
// (the block of the KeyperSetAdded event that revealed the gap). All RPC calls
// use opts derived from that triggering event. RPC failures and zero addresses
// are logged and tolerated so that one unreadable keyper set does not abort
// backfill of the rest or processing of the triggering event.
func (s *DKGSyncer) backfillKeyperSet(ctx context.Context, opts *bind.CallOpts, ksi, startBlock uint64) {
	ksAddr, err := s.getKeyperSetAddress(opts, ksi)
	if err != nil {
		s.Log.Error(
			"could not read keyper set address while backfilling gap",
			"error", err.Error(),
			"keyper-set-index", ksi,
		)
		return
	}
	dkgAddr, err := s.resolveDKGContract(ctx, opts, ksAddr)
	if err != nil {
		s.Log.Error(
			"could not resolve DKG contract while backfilling gap",
			"error", err.Error(),
			"keyper-set", ksAddr.Hex(),
			"keyper-set-index", ksi,
		)
		return
	}
	if !s.tryTrack(dkgAddr, ksi) {
		return
	}
	if s.runner == nil || s.bindDKGContract == nil {
		return
	}
	if err := s.startContractSyncer(ctx, s.runner, dkgAddr, startBlock); err != nil {
		s.Log.Error(
			"could not start DKG contract subscription while backfilling gap",
			"error", err.Error(),
			"dkg-contract", dkgAddr.Hex(),
			"keyper-set-index", ksi,
		)
	}
}

// defaultDKGContractResolver binds the KeyperSet contract at keyperSetAddr and
// calls its getDKGContract() view. This is the production resolver; tests
// override DKGSyncer.resolveDKGContract with an in-memory stub.
func defaultDKGContractResolver(backend bind.ContractBackend) dkgContractResolver {
	return func(_ context.Context, opts *bind.CallOpts, keyperSetAddr common.Address) (common.Address, error) {
		ks, err := keypersetBindings.NewKeyperset(keyperSetAddr, backend)
		if err != nil {
			return common.Address{}, errors.Wrap(err, "bind KeyperSet contract")
		}
		return ks.GetDKGContract(opts)
	}
}

// defaultDKGContractBinder binds a *contract.DKGContract at the given address.
// This is the production binder; tests override DKGSyncer.bindDKGContract
// with an in-memory fake that can emit canned events.
func defaultDKGContractBinder(backend bind.ContractBackend) dkgContractBinder {
	return func(addr common.Address) (dkgContractBackend, error) {
		return contract.NewDKGContract(addr, backend)
	}
}
