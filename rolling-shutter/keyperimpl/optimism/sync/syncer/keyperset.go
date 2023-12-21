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

type KeyperSetSyncer struct {
	Client     client.Client
	Contract   *bindings.KeyperSetManager
	Log        log.Logger
	StartBlock *uint64
	Handler    event.KeyperSetHandler

	keyperAddedCh    chan *bindings.KeyperSetManagerKeyperSetAdded
	handlerScheduler chan *event.KeyperSet
}

func (s *KeyperSetSyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}

	watchOpts := &bind.WatchOpts{
		Start:   s.StartBlock, // nil means latest
		Context: ctx,
	}
	s.keyperAddedCh = make(chan *bindings.KeyperSetManagerKeyperSetAdded, 10)
	subs, err := s.Contract.WatchKeyperSetAdded(watchOpts, s.keyperAddedCh)
	// FIXME: what to do on subs.Error()
	if err != nil {
		return err
	}
	runner.Defer(subs.Unsubscribe)
	runner.Defer(func() {
		close(s.keyperAddedCh)
	})
	runner.Go(func() error {
		return s.watchNewKeypersService(ctx)
	})
	return nil
}

func (s *KeyperSetSyncer) GetKeyperSetForBlock(ctx context.Context, opts *bind.CallOpts, b *number.BlockNumber) (*event.KeyperSet, error) {
	if b.Equal(number.LatestBlock) {
		latestBlock, err := s.Client.BlockNumber(ctx)
		if err != nil {
			return nil, err
		}
		b = number.NewBlockNumber()
		b.SetInt64(int64(latestBlock))
	}
	if opts == nil {
		opts = &bind.CallOpts{
			Context: ctx,
		}
	}
	idx, err := s.Contract.GetKeyperSetIndexByBlock(opts, b.Uint64())
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve keyper set index")
	}
	actBl, err := s.Contract.GetKeyperSetActivationBlock(opts, idx)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve keyper set activation block")
	}
	addr, err := s.Contract.GetKeyperSetAddress(opts, idx)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve keyper set address")
	}
	return s.newEvent(ctx, opts, addr, actBl)
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
