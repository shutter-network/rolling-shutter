package syncer

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
	subs, err := s.Contract.WatchKeyperSetAdded(watchOpts, s.keyperAddedCh)
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
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	if err := guardCallOpts(opts, false); err != nil {
		return nil, err
	}
	bn := s.StartBlock.ToUInt64Ptr()
	if bn == nil {
		// this should not be the case
		return nil, errors.New("start block is 'latest'")
	}

	initialKeyperSets := []*event.KeyperSet{}
	// this blocknumber specifies the argument to the contract
	// getter
	ks, err := s.GetKeyperSetForBlock(ctx, opts, s.StartBlock)
	if err != nil {
		return nil, err
	}
	initialKeyperSets = append(initialKeyperSets, ks)

	numKS, err := s.Contract.GetNumKeyperSets(opts)
	if err != nil {
		return nil, err
	}

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
	opts, _, err := fixCallOpts(ctx, s.Client, opts)
	if err != nil {
		return nil, err
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
	var atBlock uint64
	var err error

	opts, latestFromFix, err := fixCallOpts(ctx, s.Client, opts)
	if err != nil {
		return nil, err
	}

	if b.Equal(number.LatestBlock) {
		if latestFromFix == nil {
			atBlock, err = s.Client.BlockNumber(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "get current block-number")
			}
		} else {
			atBlock = *latestFromFix
		}
	} else {
		atBlock = b.Uint64()
	}

	idx, err := s.Contract.GetKeyperSetIndexByBlock(opts, atBlock)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve keyper set index at block %d", atBlock)
	}
	return s.GetKeyperSetByIndex(ctx, opts, idx)
}

func (s *KeyperSetSyncer) newEvent(
	_ context.Context,
	opts *bind.CallOpts,
	keyperSetContract common.Address,
	activationBlock uint64,
) (*event.KeyperSet, error) {
	if err := guardCallOpts(opts, false); err != nil {
		return nil, err
	}
	ks, err := bindings.NewKeyperSet(keyperSetContract, s.Client)
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
	eon, err := s.Contract.GetKeyperSetIndexByBlock(opts, activationBlock)
	if err != nil {
		return nil, makeCallError("KeyperSetIndexByBlock", err)
	}
	return &event.KeyperSet{
		ActivationBlock: activationBlock,
		Members:         members,
		Threshold:       threshold,
		Eon:             eon,
		AtBlockNumber:   number.BigToBlockNumber(opts.BlockNumber),
	}, nil
}

func (s *KeyperSetSyncer) watchNewKeypersService(ctx context.Context) error {
	for {
		select {
		case newKeypers, ok := <-s.keyperAddedCh:
			if !ok {
				return nil
			}
			opts := logToCallOpts(ctx, &newKeypers.Raw)
			newKeyperSet, err := s.newEvent(
				ctx,
				opts,
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
