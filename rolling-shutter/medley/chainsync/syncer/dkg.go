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

	// runner is captured from Start() so that handleKeyperSetAdded can spawn
	// new per-contract subscription goroutines as new DKG contract addresses
	// are discovered at runtime.
	runner service.Runner

	keyperSetAddedCh chan *bindings.KeyperSetManagerKeyperSetAdded

	trackedMu           sync.Mutex
	trackedDKGContracts map[common.Address]struct{}
	// knownKeyperSetCount is the highest known number of keyper sets
	// registered in KeyperSetManager. Used to bound the Succeeded(ksi)
	// queries when a new DKG contract subscription is started.
	knownKeyperSetCount uint64
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
	s.runner = runner

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

// scanInitialDKGContracts iterates every keyper set known to the manager and
// records the DKG contract address each one points at. Failures to read an
// individual keyper set's DKG contract (RPC error) are surfaced; zero addresses
// are logged and skipped, mirroring the runtime behaviour for KeyperSetAdded.
// As a side effect, knownKeyperSetCount is updated to the number of keyper
// sets visited.
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
	s.updateKnownKeyperSetCount(numKS)
	for i := uint64(0); i < numKS; i++ {
		ksAddr, err := indexer.GetKeyperSetAddress(opts, i)
		if err != nil {
			return errors.Wrapf(err, "get keyper set address %d", i)
		}
		dkgAddr, err := s.resolveDKGContract(ctx, opts, ksAddr)
		if err != nil {
			return errors.Wrapf(err, "resolve DKG contract for keyper set %d (%s)", i, ksAddr.Hex())
		}
		s.recordDKGContract(dkgAddr, i, ksAddr)
	}
	return nil
}

// recordDKGContract inserts a discovered DKG contract address into the
// internal tracked set. It returns true when the address was newly added.
// A zero address is rejected with a warning; an already-tracked address is a
// silent no-op (returns false). The keyper set index and address are included
// in log lines so operators can correlate warnings with on-chain state.
func (s *DKGSyncer) recordDKGContract(addr common.Address, keyperSetIndex uint64, keyperSetAddr common.Address) bool {
	if (addr == common.Address{}) {
		s.Log.Warn(
			"keyper set has no DKG contract configured; skipping",
			"keyper-set-index", keyperSetIndex,
			"keyper-set", keyperSetAddr.Hex(),
		)
		return false
	}
	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()
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

// updateKnownKeyperSetCount raises knownKeyperSetCount to at least n. The
// counter is monotonically non-decreasing because keyper set indices in
// KeyperSetManager are append-only.
func (s *DKGSyncer) updateKnownKeyperSetCount(n uint64) {
	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()
	if n > s.knownKeyperSetCount {
		s.knownKeyperSetCount = n
	}
}

func (s *DKGSyncer) getKnownKeyperSetCount() uint64 {
	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()
	return s.knownKeyperSetCount
}

// startContractSyncer delivers the initial success state for the given DKG
// contract and then starts a DKGContractSyncer to watch it for live events.
// It first queries Succeeded(ksi) for every known keyper set index at
// startBlock and delivers synthetic SuccessEvents for already-completed
// instances directly via the shared Handler. It then constructs a
// DKGContractSyncer bound to the same backend and starts it via
// runner.StartService, which sets up the five bulletin-board subscriptions from
// startBlock and forwards live events to the same Handler.
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
	for _, ev := range initial {
		if err := s.Handler(ctx, ev); err != nil {
			s.Log.Error(
				"handler for initial DKG success errored",
				"error", err.Error(),
				"dkg-contract", addr.Hex(),
			)
		}
	}

	contractSyncer := &DKGContractSyncer{
		Addr:       addr,
		Log:        s.Log,
		Handler:    s.Handler,
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
	numKS := s.getKnownKeyperSetCount()
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

// handleKeyperSetAdded resolves the DKG contract address for the newly-added
// keyper set, records it in the tracked set, and -- if the address is newly
// added and a runner is available -- spawns a new per-contract subscription
// goroutine starting from the block of the triggering event. RPC failures
// and zero addresses are logged but do not abort the event loop.
func (s *DKGSyncer) handleKeyperSetAdded(ctx context.Context, ev *bindings.KeyperSetManagerKeyperSetAdded) {
	s.updateKnownKeyperSetCount(ev.Eon + 1)

	opts := logToCallOpts(ctx, &ev.Raw)
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
	newlyAdded := s.recordDKGContract(dkgAddr, ev.Eon, ev.KeyperSetContract)
	if !newlyAdded {
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
