package syncer

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/contract"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

type ECIESKeySyncer struct {
	Client     client.Client
	Contract   *contract.ECIESKeyRegistry
	Log        log.Logger
	StartBlock *number.BlockNumber
	Handler    event.ECIESKeyHandler

	keyRegisteredCh chan *contract.ECIESKeyRegistryKeyRegistered
}

func (s *ECIESKeySyncer) Start(ctx context.Context, runner service.Runner) error {
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
	initial, err := s.getInitialKeys(ctx)
	if err != nil {
		return err
	}
	for _, k := range initial {
		err = s.Handler(ctx, k)
		if err != nil {
			s.Log.Error(
				"handler for `ECIESKey` errored for initial sync",
				"error",
				err.Error(),
			)
		}
	}

	s.keyRegisteredCh = make(chan *contract.ECIESKeyRegistryKeyRegistered, channelSize)
	subs, err := s.Contract.WatchKeyRegistered(watchOpts, s.keyRegisteredCh, nil)
	if err != nil {
		return err
	}
	runner.Go(func() error {
		err := s.watchNewKeys(ctx, subs.Err())
		if err != nil {
			s.Log.Error("error watching new ECIES keys", err.Error())
		}
		subs.Unsubscribe()
		return err
	})
	return nil
}

// getInitialKeys iterates the registry's own keyper list and emits a synthetic
// ECIESKey event for every keyper whose key is already registered, so the
// local cache is populated before any live events arrive.
func (s *ECIESKeySyncer) getInitialKeys(ctx context.Context) ([]*event.ECIESKey, error) {
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	if err := guardCallOpts(opts, false); err != nil {
		return nil, err
	}

	count, err := s.Contract.GetKeyperCount(opts)
	if err != nil {
		return nil, errors.Wrap(err, "get keyper count")
	}

	keys := []*event.ECIESKey{}
	total := count.Uint64()
	for i := uint64(0); i < total; i++ {
		addr, err := s.Contract.GetKeyperAt(opts, new(big.Int).SetUint64(i))
		if err != nil {
			return nil, errors.Wrapf(err, "get keyper at index %d", i)
		}
		key, err := s.Contract.GetKey(opts, addr)
		if err != nil {
			return nil, errors.Wrapf(err, "get ECIES key for %s", addr.Hex())
		}
		if len(key) == 0 {
			continue
		}
		keys = append(keys, &event.ECIESKey{
			Keyper:         addr,
			EciesPublicKey: key,
			AtBlockNumber:  number.BigToBlockNumber(opts.BlockNumber),
		})
	}
	return keys, nil
}

func (s *ECIESKeySyncer) watchNewKeys(ctx context.Context, subsErr <-chan error) error {
	for {
		select {
		case ev, ok := <-s.keyRegisteredCh:
			if !ok {
				return nil
			}
			bn := ev.Raw.BlockNumber
			err := s.Handler(ctx, &event.ECIESKey{
				Keyper:         ev.Keyper,
				EciesPublicKey: ev.EciesPublicKey,
				AtBlockNumber:  number.NewBlockNumber(&bn),
			})
			if err != nil {
				s.Log.Error(
					"handler for `ECIESKey` errored",
					"error",
					err.Error(),
				)
			}
		case err := <-subsErr:
			if err != nil {
				s.Log.Error("subscription error for watchNewKeys", err.Error())
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
