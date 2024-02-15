package syncer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

var _ ManualFilterHandler = &EonPubKeySyncer{}

type EonPubKeySyncer struct {
	Client              client.EthereumClient
	Log                 log.Logger
	KeyBroadcast        *bindings.KeyBroadcastContract
	KeyperSetManager    *bindings.KeyperSetManager
	StartBlock          *number.BlockNumber
	Handler             event.EonPublicKeyHandler
	DisableEventWatcher bool

	keyBroadcastCh chan *bindings.KeyBroadcastContractEonKeyBroadcast
}

func (s *EonPubKeySyncer) QueryAndHandle(ctx context.Context, block uint64) error {
	s.Log.Info(
		"pubsyncer query and handle called",
		"block",
		block,
	)
	opts := &bind.FilterOpts{
		Start:   block,
		End:     &block,
		Context: ctx,
	}
	iter, err := s.KeyBroadcast.FilterEonKeyBroadcast(opts)
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.Next() {
		select {
		case s.keyBroadcastCh <- iter.Event:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err := iter.Error(); err != nil {
		return errors.Wrap(err, "filter iterator error")
	}
	return nil
}

func (s *EonPubKeySyncer) Start(ctx context.Context, runner service.Runner) error {
	fmt.Println("pubsync looper started (println)")
	s.Log.Info(
		"pubsyncer loop started",
	)
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
	s.keyBroadcastCh = make(chan *bindings.KeyBroadcastContractEonKeyBroadcast, channelSize)
	runner.Defer(func() {
		close(s.keyBroadcastCh)
	})
	if !s.DisableEventWatcher {
		subs, err := s.KeyBroadcast.WatchEonKeyBroadcast(watchOpts, s.keyBroadcastCh)
		// FIXME: what to do on subs.Error()
		if err != nil {
			return err
		}
		runner.Defer(subs.Unsubscribe)
	}
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
			s.Log.Info(
				"pubsyncer received value",
			)
			if !ok {
				return nil
			}
			s.Log.Info(
				"pubsyncer channel ok",
			)
			// FIXME: this happens, why?
			if len(newEonKey.Key) == 0 {
				opts := &bind.CallOpts{
					Context:     ctx,
					BlockNumber: new(big.Int).SetUint64(newEonKey.Raw.BlockNumber),
				}
				k, err := s.GetEonPubKeyForEon(ctx, opts, newEonKey.Eon)
				s.Log.Error(
					"extra call for GetEonPubKeyForEon errored",
					"error",
					err.Error(),
				)
				s.Log.Info(
					"retrieved eon pubkey by getter",
					"eon",
					k,
				)
			} else {
				s.Log.Info(
					"pubsyncer key lenght ok",
				)
			}
			pubk := newEonKey.Key
			bn := newEonKey.Raw.BlockNumber
			ev := &event.EonPublicKey{
				Eon:           newEonKey.Eon,
				Key:           pubk,
				AtBlockNumber: number.NewBlockNumber(&bn),
			}
			s.Log.Info(
				"pubsyncer constructed event",
				"event",
				ev,
			)
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
