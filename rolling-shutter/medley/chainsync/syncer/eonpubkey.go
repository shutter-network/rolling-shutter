package syncer

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shutter-network/contracts/v2/bindings/keybroadcastcontract"
	"github.com/shutter-network/contracts/v2/bindings/keypersetmanager"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type EonPubKeySyncer struct {
	Client           client.Client
	Log              log.Logger
	KeyBroadcast     *keybroadcastcontract.Keybroadcastcontract
	KeyperSetManager *keypersetmanager.Keypersetmanager
	StartBlock       *number.BlockNumber
	Handler          event.EonPublicKeyHandler

	keyBroadcastCh chan *keybroadcastcontract.KeybroadcastcontractEonKeyBroadcast
}

func (s *EonPubKeySyncer) Start(ctx context.Context, runner service.Runner) error {
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
	pubKs, err := s.getInitialPubKeys(ctx)
	if err != nil {
		return err
	}
	for _, k := range pubKs {
		err := s.Handler(ctx, k)
		if err != nil {
			return err
		}
	}

	watchOpts := &bind.WatchOpts{
		Start:   s.StartBlock.ToUInt64Ptr(),
		Context: ctx,
	}
	s.keyBroadcastCh = make(chan *keybroadcastcontract.KeybroadcastcontractEonKeyBroadcast, channelSize)
	subs, err := s.KeyBroadcast.WatchEonKeyBroadcast(watchOpts, s.keyBroadcastCh)
	// FIXME: what to do on subs.Error()
	if err != nil {
		return err
	}
	runner.Defer(subs.Unsubscribe)
	runner.Go(func() error {
		return s.watchNewEonPubkey(ctx)
	})
	return nil
}

func (s *EonPubKeySyncer) getInitialPubKeys(ctx context.Context) ([]*event.EonPublicKey, error) {
	// This blocknumber specifies AT what state
	// the contract is called
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	numKS, err := s.KeyperSetManager.GetNumKeyperSets(opts)
	if err != nil {
		return nil, err
	}
	// this blocknumber specifies the argument to the contract
	// getter
	activeEon, err := s.KeyperSetManager.GetKeyperSetIndexByBlock(opts, s.StartBlock.Uint64())
	if err != nil {
		return nil, err
	}

	initialPubKeys := []*event.EonPublicKey{}
	for i := activeEon; i < numKS; i++ {
		e, err := s.GetEonPubKeyForEon(ctx, opts, i)
		// FIXME: translate the error that there is no key
		// to a continue of the loop
		// (key not in mapping error, how can we catch that?)
		if err != nil {
			return nil, err
		}
		initialPubKeys = append(initialPubKeys, e)
	}
	return initialPubKeys, nil
}

func (s *EonPubKeySyncer) logCallError(attrName string, err error) {
	s.Log.Error(
		fmt.Sprintf("could not retrieve `%s` from contract", attrName),
		"error",
		err.Error(),
	)
}

func (s *EonPubKeySyncer) GetEonPubKeyForEon(ctx context.Context, opts *bind.CallOpts, eon uint64) (*event.EonPublicKey, error) {
	var err error
	opts, _, err = fixCallOpts(ctx, s.Client, opts)
	if err != nil {
		return nil, err
	}
	key, err := s.KeyBroadcast.GetEonKey(opts, eon)
	// XXX: can the key be a null byte?
	// I think we rather get a index out of bounds error.
	if err != nil {
		return nil, err
	}
	return &event.EonPublicKey{
		Eon:           eon,
		Key:           key,
		AtBlockNumber: number.BigToBlockNumber(opts.BlockNumber),
	}, nil
}

func (s *EonPubKeySyncer) watchNewEonPubkey(ctx context.Context) error {
	for {
		select {
		case newEonKey, ok := <-s.keyBroadcastCh:
			if !ok {
				return nil
			}
			bn := newEonKey.Raw.BlockNumber
			ev := &event.EonPublicKey{
				Eon:           newEonKey.Eon,
				Key:           newEonKey.Key,
				AtBlockNumber: number.NewBlockNumber(&bn),
			}
			err := s.Handler(ctx, ev)
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
