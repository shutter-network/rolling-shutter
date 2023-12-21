package syncer

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type EonPubKeySyncer struct {
	Client     client.Client
	Log        log.Logger
	Contract   *bindings.KeyBroadcastContract
	StartBlock *uint64
	Handler    event.EonPublicKeyHandler

	keyBroadcastCh chan *bindings.KeyBroadcastContractEonKeyBroadcast
}

func (s *EonPubKeySyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}
	watchOpts := &bind.WatchOpts{
		Start:   s.StartBlock, // nil means latest
		Context: ctx,
	}
	s.keyBroadcastCh = make(chan *bindings.KeyBroadcastContractEonKeyBroadcast, 10)
	subs, err := s.Contract.WatchEonKeyBroadcast(watchOpts, s.keyBroadcastCh)
	// FIXME: what to do on subs.Error()
	if err != nil {
		return err
	}
	runner.Defer(subs.Unsubscribe)
	runner.Defer(func() {
		close(s.keyBroadcastCh)
	})
	runner.Go(func() error {
		return s.watchNewEonPubkey(ctx)
	})
	return nil
}

func (s *EonPubKeySyncer) logCallError(attrName string, err error) {
	s.Log.Error(
		fmt.Sprintf("could not retrieve `%s` from contract", attrName),
		"error",
		err.Error(),
	)
}

func (s *EonPubKeySyncer) GetEonPubKeyForEon(ctx context.Context, opts *bind.CallOpts, eon uint64) (*event.EonPublicKey, error) {
	if opts == nil {
		opts = &bind.CallOpts{
			Context: ctx,
		}
	}
	key, err := s.Contract.GetEonKey(opts, eon)
	// XXX: can the key be a null byte?
	// I think we rather get a index out of bounds error.
	if err != nil {
		return nil, err
	}
	return &event.EonPublicKey{
		Eon: eon,
		Key: key,
	}, nil
}

func (s *EonPubKeySyncer) watchNewEonPubkey(ctx context.Context) error {
	for {
		select {
		case newEonKey, ok := <-s.keyBroadcastCh:
			if !ok {
				return nil
			}
			ev := &event.EonPublicKey{
				Eon: newEonKey.Eon,
				Key: newEonKey.Key,
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
