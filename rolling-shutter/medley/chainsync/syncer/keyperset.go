package syncer

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethevent "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func makeCallError(attrName string, err error) error {
	return errors.Wrapf(err, "could not retrieve `%s` from contract", attrName)
}

const channelSize = 10

// keyperSetManager is the subset of *bindings.KeyperSetManager the
// KeyperSetSyncer calls: the by-index enumeration reads and the KeyperSetAdded
// watch. Pulling it out as an interface lets tests substitute a fake manager
// with canned keyper sets and events without spinning up a simulated chain
// (consistent with ADR 0007). Note the absence of GetKeyperSetIndexByBlock:
// the syncer addresses keyper sets by index and never derives one from a block.
type keyperSetManager interface {
	GetNumKeyperSets(opts *bind.CallOpts) (uint64, error)
	GetKeyperSetAddress(opts *bind.CallOpts, index uint64) (common.Address, error)
	GetKeyperSetActivationBlock(opts *bind.CallOpts, index uint64) (uint64, error)
	WatchKeyperSetAdded(
		opts *bind.WatchOpts,
		sink chan<- *bindings.KeyperSetManagerKeyperSetAdded,
	) (gethevent.Subscription, error)
}

// keyperSetReader is the subset of a single deployed *bindings.KeyperSet
// contract the syncer reads when building a KeyperSet event. Injectable via
// keyperSetBinder so tests need not bind a real contract.
type keyperSetReader interface {
	IsFinalized(opts *bind.CallOpts) (bool, error)
	GetMembers(opts *bind.CallOpts) ([]common.Address, error)
	GetThreshold(opts *bind.CallOpts) (uint64, error)
}

// keyperSetBinder constructs a keyperSetReader for a KeyperSet contract address.
// The production implementation binds a *bindings.KeyperSet; tests substitute an
// in-memory fake.
type keyperSetBinder func(addr common.Address) (keyperSetReader, error)

type KeyperSetSyncer struct {
	Client     client.Client
	Contract   *bindings.KeyperSetManager
	Log        log.Logger
	StartBlock *number.BlockNumber
	Handler    event.KeyperSetHandler

	// manager is the source of manager-level reads and the KeyperSetAdded watch.
	// Overridable for tests; production callers leave it nil and Start() fills in
	// the bound Contract.
	manager keyperSetManager
	// bindKeyperSet constructs a keyperSetReader for a KeyperSet contract
	// address. Overridable for tests; production callers leave it nil and Start()
	// fills in the default binding-backed implementation over Client.
	bindKeyperSet keyperSetBinder

	keyperAddedCh chan *bindings.KeyperSetManagerKeyperSetAdded
}

// defaultKeyperSetBinder binds a *bindings.KeyperSet at the given address. This
// is the production binder; tests override KeyperSetSyncer.bindKeyperSet with an
// in-memory fake.
func defaultKeyperSetBinder(backend bind.ContractBackend) keyperSetBinder {
	return func(addr common.Address) (keyperSetReader, error) {
		return bindings.NewKeyperSet(addr, backend)
	}
}

func (s *KeyperSetSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}

	if s.manager == nil {
		s.manager = s.Contract
	}
	if s.bindKeyperSet == nil {
		s.bindKeyperSet = defaultKeyperSetBinder(s.Client)
	}

	// the latest block still has to be fixed.
	// otherwise we could skip some block events
	// between the initial poll and the subscription.
	if s.StartBlock.IsLatest() {
		latest, err := s.Client.BlockNumber(ctx)
		if err != nil {
			return err
		}
		s.StartBlock.SetUint64(latest)
	}

	watchOpts := &bind.WatchOpts{
		Start:   s.StartBlock.ToUInt64Ptr(),
		Context: ctx,
	}
	initial, err := s.getInitialKeyperSets(ctx)
	if err != nil {
		return err
	}
	for _, ks := range initial {
		err = s.Handler(ctx, ks)
		if err != nil {
			s.Log.Error(
				"handler for `NewKeyperSet` errored for initial sync",
				"error",
				err.Error(),
			)
		}
	}
	s.keyperAddedCh = make(chan *bindings.KeyperSetManagerKeyperSetAdded, channelSize)
	subs, err := s.manager.WatchKeyperSetAdded(watchOpts, s.keyperAddedCh)
	if err != nil {
		return err
	}
	runner.Go(func() error {
		err := s.watchNewKeypersService(ctx, subs.Err())
		if err != nil {
			s.Log.Error("error watching new keypers", err.Error())
		}
		subs.Unsubscribe()
		return err
	})
	return nil
}

// getInitialKeyperSets performs the cold-start poll. It enumerates every
// registered keyper set by index (0..GetNumKeyperSets()-1) and delivers each to
// the handler, rather than asking which set is active at the start block. This
// backfills the full history that the KeyperSetAdded watch — which only ever
// delivers events at or after the start block — can never replay, and starts
// cleanly when no set is active yet (an empty contract yields nothing). Keyper
// sets are addressed by index; activation is a downstream scheduling concern.
// See ADR 0010.
func (s *KeyperSetSyncer) getInitialKeyperSets(ctx context.Context) ([]*event.KeyperSet, error) {
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	if err := guardCallOpts(opts, false); err != nil {
		return nil, err
	}

	numKS, err := s.manager.GetNumKeyperSets(opts)
	if err != nil {
		return nil, err
	}

	initialKeyperSets := make([]*event.KeyperSet, 0, numKS)
	for i := uint64(0); i < numKS; i++ {
		ks, err := s.GetKeyperSetByIndex(ctx, opts, i)
		if err != nil {
			return nil, err
		}
		initialKeyperSets = append(initialKeyperSets, ks)
	}

	return initialKeyperSets, nil
}

func (s *KeyperSetSyncer) GetKeyperSetByIndex(ctx context.Context, opts *bind.CallOpts, index uint64) (*event.KeyperSet, error) {
	opts, err := fixCallOpts(ctx, s.Client, opts)
	if err != nil {
		return nil, err
	}
	actBl, err := s.manager.GetKeyperSetActivationBlock(opts, index)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve keyper set activation block")
	}
	addr, err := s.manager.GetKeyperSetAddress(opts, index)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve keyper set address")
	}
	return s.newEvent(opts, addr, actBl, index)
}

// newEvent builds a KeyperSet event from the per-set contract reads. The
// keyperSetIndex is passed in by the caller — the initial poll knows it from the
// enumeration loop and the watch knows it from the event's Eon field — so no
// call derives an index from an activation block.
func (s *KeyperSetSyncer) newEvent(
	opts *bind.CallOpts,
	keyperSetContract common.Address,
	activationBlock uint64,
	keyperSetIndex uint64,
) (*event.KeyperSet, error) {
	if err := guardCallOpts(opts, false); err != nil {
		return nil, err
	}
	ks, err := s.bindKeyperSet(keyperSetContract)
	if err != nil {
		return nil, errors.Wrap(err, "could not bind to KeyperSet contract")
	}
	// the manager only accepts final keyper sets,
	// so we expect this to be final now.
	final, err := ks.IsFinalized(opts)
	if err != nil {
		return nil, makeCallError("IsFinalized", err)
	}
	if !final {
		return nil, errors.New("contract did accept unfinalized keyper-sets")
	}
	members, err := ks.GetMembers(opts)
	if err != nil {
		return nil, makeCallError("Members", err)
	}
	threshold, err := ks.GetThreshold(opts)
	if err != nil {
		return nil, makeCallError("Threshold", err)
	}
	return &event.KeyperSet{
		ActivationBlock: activationBlock,
		Members:         members,
		Threshold:       threshold,
		Eon:             keyperSetIndex,
		Contract:        keyperSetContract,
		AtBlockNumber:   number.BigToBlockNumber(opts.BlockNumber),
	}, nil
}

func (s *KeyperSetSyncer) watchNewKeypersService(ctx context.Context, subsErr <-chan error) error {
	for {
		select {
		case newKeypers, ok := <-s.keyperAddedCh:
			if !ok {
				return nil
			}
			opts := logToCallOpts(ctx, &newKeypers.Raw)
			newKeyperSet, err := s.newEvent(
				opts,
				newKeypers.KeyperSetContract,
				newKeypers.ActivationBlock,
				newKeypers.Eon,
			)
			if err != nil {
				s.Log.Error(
					"error while fetching new event",
					"error",
					err.Error(),
				)
				continue
			}
			err = s.Handler(ctx, newKeyperSet)
			if err != nil {
				s.Log.Error(
					"handler for `NewKeyperSet` errored",
					"error",
					err.Error(),
				)
			}
		case err := <-subsErr:
			if err != nil {
				s.Log.Error("subscription error for watchNewKeypersService", err.Error())
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
