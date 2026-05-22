package syncer

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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

// DKGEventSyncer mirrors the structure of KeyperSetSyncer: it performs an
// initial poll for already-succeeded DKG Instances and then subscribes to the
// five bulletin-board event types emitted by the DKG Contract.
//
// In addition to its single statically-configured DKG contract, the syncer
// also discovers DKG contract addresses autonomously by scanning every keyper
// set in KeyperSetManager and subscribing to KeyperSetAdded. The discovered
// addresses are deduplicated in trackedDKGContracts; this set is the basis on
// which per-contract event subscriptions are wired up in subsequent slices.
type DKGEventSyncer struct {
	Client           client.Client
	Contract         *contract.DKGContract
	KeyperSetManager *bindings.KeyperSetManager
	Log              log.Logger
	StartBlock       *number.BlockNumber
	Handler          event.DKGEventHandler

	dealingCh     chan *contract.DKGContractDealingSubmitted
	accusationCh  chan *contract.DKGContractAccusationSubmitted
	apologyCh     chan *contract.DKGContractApologySubmitted
	successVoteCh chan *contract.DKGContractSuccessVoteSubmitted
	successCh     chan *contract.DKGContractDKGSucceeded

	// resolveDKGContract is the function used to read the DKG contract address
	// from a KeyperSet contract. It is overridable so tests can avoid binding
	// to a real contract; production callers leave it nil and Start() fills in
	// the default binding-backed implementation.
	resolveDKGContract dkgContractResolver

	keyperSetAddedCh chan *bindings.KeyperSetManagerKeyperSetAdded

	trackedMu           sync.Mutex
	trackedDKGContracts map[common.Address]struct{}
}

func (s *DKGEventSyncer) Start(ctx context.Context, runner service.Runner) error {
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

	watchOpts := &bind.WatchOpts{
		Start:   s.StartBlock.ToUInt64Ptr(),
		Context: ctx,
	}

	callOpts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	if err := s.scanInitialDKGContracts(ctx, callOpts, s.KeyperSetManager); err != nil {
		return errors.Wrap(err, "initial DKG contract scan")
	}

	initial, err := s.getInitialSuccesses(ctx)
	if err != nil {
		return err
	}
	for _, ev := range initial {
		if err := s.Handler(ctx, ev); err != nil {
			s.Log.Error(
				"handler for initial DKG success errored",
				"error",
				err.Error(),
			)
		}
	}

	s.dealingCh = make(chan *contract.DKGContractDealingSubmitted, channelSize)
	s.accusationCh = make(chan *contract.DKGContractAccusationSubmitted, channelSize)
	s.apologyCh = make(chan *contract.DKGContractApologySubmitted, channelSize)
	s.successVoteCh = make(chan *contract.DKGContractSuccessVoteSubmitted, channelSize)
	s.successCh = make(chan *contract.DKGContractDKGSucceeded, channelSize)

	dealingSub, err := s.Contract.WatchDealingSubmitted(watchOpts, s.dealingCh, nil, nil, nil)
	if err != nil {
		return errors.Wrap(err, "watch DealingSubmitted")
	}
	accusationSub, err := s.Contract.WatchAccusationSubmitted(watchOpts, s.accusationCh, nil, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		return errors.Wrap(err, "watch AccusationSubmitted")
	}
	apologySub, err := s.Contract.WatchApologySubmitted(watchOpts, s.apologyCh, nil, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		return errors.Wrap(err, "watch ApologySubmitted")
	}
	successVoteSub, err := s.Contract.WatchSuccessVoteSubmitted(watchOpts, s.successVoteCh, nil, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		apologySub.Unsubscribe()
		return errors.Wrap(err, "watch SuccessVoteSubmitted")
	}
	successSub, err := s.Contract.WatchDKGSucceeded(watchOpts, s.successCh, nil, nil)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		apologySub.Unsubscribe()
		successVoteSub.Unsubscribe()
		return errors.Wrap(err, "watch DKGSucceeded")
	}

	s.keyperSetAddedCh = make(chan *bindings.KeyperSetManagerKeyperSetAdded, channelSize)
	keyperAddedSub, err := s.KeyperSetManager.WatchKeyperSetAdded(watchOpts, s.keyperSetAddedCh)
	if err != nil {
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		apologySub.Unsubscribe()
		successVoteSub.Unsubscribe()
		successSub.Unsubscribe()
		return errors.Wrap(err, "watch KeyperSetAdded")
	}

	runner.Go(func() error {
		err := s.watchEvents(
			ctx,
			dealingSub.Err(),
			accusationSub.Err(),
			apologySub.Err(),
			successVoteSub.Err(),
			successSub.Err(),
			keyperAddedSub.Err(),
		)
		if err != nil {
			s.Log.Error("error watching DKG events", err.Error())
		}
		dealingSub.Unsubscribe()
		accusationSub.Unsubscribe()
		apologySub.Unsubscribe()
		successVoteSub.Unsubscribe()
		successSub.Unsubscribe()
		keyperAddedSub.Unsubscribe()
		return err
	})
	return nil
}

// scanInitialDKGContracts iterates every keyper set known to the manager and
// records the DKG contract address each one points at. Failures to read an
// individual keyper set's DKG contract (RPC error) are surfaced; zero addresses
// are logged and skipped, mirroring the runtime behaviour for KeyperSetAdded.
func (s *DKGEventSyncer) scanInitialDKGContracts(
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
		s.recordDKGContract(dkgAddr, i, ksAddr)
	}
	return nil
}

// recordDKGContract inserts a discovered DKG contract address into the
// internal tracked set. It returns true when the address was newly added.
// A zero address is rejected with a warning; an already-tracked address is a
// silent no-op (returns false). The keyper set index and address are included
// in log lines so operators can correlate warnings with on-chain state.
func (s *DKGEventSyncer) recordDKGContract(addr common.Address, keyperSetIndex uint64, keyperSetAddr common.Address) bool {
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
// contract addresses. Intended for tests and for slice 002 which will iterate
// the set to spawn per-contract subscription goroutines.
func (s *DKGEventSyncer) trackedDKGContractList() []common.Address {
	s.trackedMu.Lock()
	defer s.trackedMu.Unlock()
	out := make([]common.Address, 0, len(s.trackedDKGContracts))
	for addr := range s.trackedDKGContracts {
		out = append(out, addr)
	}
	return out
}

// getInitialSuccesses queries DKGContract.succeeded(k) for each known keyper
// set index. Already-succeeded instances are delivered as synthetic events so
// the local cache (e.g. dkg_result) can be populated before any live events.
func (s *DKGEventSyncer) getInitialSuccesses(ctx context.Context) ([]event.DKGEvent, error) {
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	if err := guardCallOpts(opts, false); err != nil {
		return nil, err
	}

	numKS, err := s.KeyperSetManager.GetNumKeyperSets(opts)
	if err != nil {
		return nil, errors.Wrap(err, "get num keyper sets")
	}

	events := []event.DKGEvent{}
	for i := uint64(0); i < numKS; i++ {
		succeeded, err := s.Contract.Succeeded(opts, i)
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

func (s *DKGEventSyncer) watchEvents(
	ctx context.Context,
	dealingErr, accusationErr, apologyErr, successVoteErr, successErr, keyperAddedErr <-chan error,
) error {
	for {
		select {
		case ev, ok := <-s.dealingCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.DealingEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				Commitment:     ev.Commitment,
				PolyEvals:      ev.PolyEvals,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-s.accusationCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.AccusationEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				AccusedIndices: ev.AccusedIndices,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-s.apologyCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.ApologyEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				AccuserIndices: ev.AccuserIndices,
				PolyEvalData:   ev.PolyEvalData,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-s.successVoteCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.SuccessVoteEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				KeyperIndex:    ev.KeyperIndex,
				EonPublicKey:   ev.EonPublicKey,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-s.successCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			s.deliver(ctx, &event.SuccessEvent{
				KeyperSetIndex: ev.KeyperSetIndex,
				RetryCounter:   ev.RetryCounter,
				EonPublicKey:   ev.EonPublicKey,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
		case ev, ok := <-s.keyperSetAddedCh:
			if !ok {
				return nil
			}
			s.handleKeyperSetAdded(ctx, ev)
		case err := <-dealingErr:
			if err != nil {
				return errors.Wrap(err, "DealingSubmitted subscription")
			}
		case err := <-accusationErr:
			if err != nil {
				return errors.Wrap(err, "AccusationSubmitted subscription")
			}
		case err := <-apologyErr:
			if err != nil {
				return errors.Wrap(err, "ApologySubmitted subscription")
			}
		case err := <-successVoteErr:
			if err != nil {
				return errors.Wrap(err, "SuccessVoteSubmitted subscription")
			}
		case err := <-successErr:
			if err != nil {
				return errors.Wrap(err, "DKGSucceeded subscription")
			}
		case err := <-keyperAddedErr:
			if err != nil {
				return errors.Wrap(err, "KeyperSetAdded subscription")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handleKeyperSetAdded resolves the DKG contract address for the newly-added
// keyper set and records it in the tracked set. RPC failures and zero
// addresses are logged but do not abort the event loop — a single bad keyper
// set must not bring down DKG event syncing for the rest.
func (s *DKGEventSyncer) handleKeyperSetAdded(ctx context.Context, ev *bindings.KeyperSetManagerKeyperSetAdded) {
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
	s.recordDKGContract(dkgAddr, ev.Eon, ev.KeyperSetContract)
}

func (s *DKGEventSyncer) deliver(ctx context.Context, ev event.DKGEvent) {
	if err := s.Handler(ctx, ev); err != nil {
		s.Log.Error(
			"handler for DKG event errored",
			"error",
			err.Error(),
		)
	}
}

// defaultDKGContractResolver binds the KeyperSet contract at keyperSetAddr and
// calls its getDKGContract() view. This is the production resolver; tests
// override DKGEventSyncer.resolveDKGContract with an in-memory stub.
func defaultDKGContractResolver(backend bind.ContractBackend) dkgContractResolver {
	return func(_ context.Context, opts *bind.CallOpts, keyperSetAddr common.Address) (common.Address, error) {
		ks, err := keypersetBindings.NewKeyperset(keyperSetAddr, backend)
		if err != nil {
			return common.Address{}, errors.Wrap(err, "bind KeyperSet contract")
		}
		return ks.GetDKGContract(opts)
	}
}
