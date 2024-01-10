package syncer

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

func makeCallError(attrName string, err error) error {
	return errors.Wrapf(err, "could not retrieve `%s` from contract", attrName)
}

const channelSize = 10

type KeyperSetSyncer struct {
	Client     client.Client
	Contract   *bindings.KeyperSetManager
	Log        log.Logger
	StartBlock *number.BlockNumber
	Handler    event.KeyperSetHandler

	keyperAddedCh chan *bindings.KeyperSetManagerKeyperSetAdded
}

func (s *KeyperSetSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
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
	runner.Defer(func() {
		close(s.keyperAddedCh)
	})
	subs, err := s.Contract.WatchKeyperSetAdded(watchOpts, s.keyperAddedCh)
	// FIXME: what to do on subs.Error()
	if err != nil {
		return err
	}
	runner.Defer(subs.Unsubscribe)
	runner.Go(func() error {
		return s.watchNewKeypersService(ctx)
	})
	return nil
}

func (s *KeyperSetSyncer) getInitialKeyperSets(ctx context.Context) ([]*event.KeyperSet, error) {
	// This blocknumber specifies AT what state
	// the contract is called
	// XXX: does the call-opts blocknumber -1 also means latest?
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	numKS, err := s.Contract.GetNumKeyperSets(opts)
	if err != nil {
		return nil, err
	}

	initialKeyperSets := []*event.KeyperSet{}
	// this blocknumber specifies the argument to the contract
	// getter
	ks, err := s.GetKeyperSetForBlock(ctx, nil, s.StartBlock)
	if err != nil {
		return nil, err
	}
	initialKeyperSets = append(initialKeyperSets, ks)

	for i := ks.Eon + 1; i < numKS; i++ {
		ks, err = s.GetKeyperSetByIndex(ctx, opts, i)
		if err != nil {
			return nil, err
		}
		initialKeyperSets = append(initialKeyperSets, ks)
	}

	return initialKeyperSets, nil
}

func (s *KeyperSetSyncer) GetKeyperSetByIndex(ctx context.Context, opts *bind.CallOpts, index uint64) (*event.KeyperSet, error) {
	if opts == nil {
		opts = &bind.CallOpts{
			Context: ctx,
		}
	}
	actBl, err := s.Contract.GetKeyperSetActivationBlock(opts, index)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve keyper set activation block")
	}
	addr, err := s.Contract.GetKeyperSetAddress(opts, index)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve keyper set address")
	}
	return s.newEvent(ctx, opts, addr, actBl)
}

func (s *KeyperSetSyncer) GetKeyperSetForBlock(ctx context.Context, opts *bind.CallOpts, b *number.BlockNumber) (*event.KeyperSet, error) {
	var latestBlock uint64
	var err error

	if b.Equal(number.LatestBlock) {
		latestBlock, err = s.Client.BlockNumber(ctx)
	} else {
		latestBlock = b.Uint64()
	}
	if opts == nil {
		// call at "latest block" state
		opts = &bind.CallOpts{
			Context: ctx,
		}
	}
	idx, err := s.Contract.GetKeyperSetIndexByBlock(opts, latestBlock)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve keyper set index")
	}
	return s.GetKeyperSetByIndex(ctx, opts, idx)
}

func (s *KeyperSetSyncer) newEvent(
	ctx context.Context,
	opts *bind.CallOpts,
	keyperSetContract common.Address,
	activationBlock uint64,
) (*event.KeyperSet, error) {
	callOpts := opts
	if callOpts == nil {
		callOpts = &bind.CallOpts{
			Context: ctx,
		}
	}
	ks, err := bindings.NewKeyperSet(keyperSetContract, s.Client)
	if err != nil {
		return nil, errors.Wrap(err, "could not bind to KeyperSet contract")
	}
	// the manager only accepts final keyper sets,
	// so we expect this to be final now.
	final, err := ks.IsFinalized(callOpts)
	if err != nil {
		return nil, makeCallError("IsFinalized", err)
	}
	if !final {
		return nil, errors.New("contract did accept unfinalized keyper-sets")
	}
	members, err := ks.GetMembers(callOpts)
	if err != nil {
		return nil, makeCallError("Members", err)
	}
	threshold, err := ks.GetThreshold(callOpts)
	if err != nil {
		return nil, makeCallError("Threshold", err)
	}
	eon, err := s.Contract.GetKeyperSetIndexByBlock(callOpts, activationBlock)
	if err != nil {
		return nil, makeCallError("KeyperSetIndexByBlock", err)
	}
	return &event.KeyperSet{
		ActivationBlock: activationBlock,
		Members:         members,
		Threshold:       threshold,
		Eon:             eon,
	}, nil
}

func (s *KeyperSetSyncer) watchNewKeypersService(ctx context.Context) error {
	for {
		select {
		case newKeypers, ok := <-s.keyperAddedCh:
			if !ok {
				return nil
			}
			newKeyperSet, err := s.newEvent(
				ctx,
				nil,
				newKeypers.KeyperSetContract,
				newKeypers.ActivationBlock,
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
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
