package syncer

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethevent "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

// keyperSetCounter is the subset of *bindings.KeyperSetManager the
// EonPubKeySyncer calls: just the total count that bounds the by-index
// enumeration. The syncer addresses eon keys by keyper-set index and never
// derives one from a block, so GetKeyperSetIndexByBlock is deliberately absent.
// Pulling it out as an interface lets tests drive the enumeration with a fake
// count instead of a simulated chain (consistent with ADR 0007).
type keyperSetCounter interface {
	GetNumKeyperSets(opts *bind.CallOpts) (uint64, error)
}

// eonKeyBroadcast is the subset of *bindings.KeyBroadcastContract the syncer
// calls: the per-index key read and the EonKeyBroadcast watch. Injectable so
// tests can hand out canned keys (including the empty-bytes "not published yet"
// sentinel) and fire watch events without binding a real contract.
type eonKeyBroadcast interface {
	GetEonKey(opts *bind.CallOpts, eon uint64) ([]byte, error)
	WatchEonKeyBroadcast(
		opts *bind.WatchOpts,
		sink chan<- *bindings.KeyBroadcastContractEonKeyBroadcast,
	) (gethevent.Subscription, error)
}

type EonPubKeySyncer struct {
	Client           client.Client
	Log              log.Logger
	KeyBroadcast     *bindings.KeyBroadcastContract
	KeyperSetManager *bindings.KeyperSetManager
	StartBlock       *number.BlockNumber
	Handler          event.EonPublicKeyHandler

	// manager sources the keyper-set count that bounds the cold-start
	// enumeration. keyBroadcast sources the per-index key reads and the
	// EonKeyBroadcast watch. Both are overridable for tests; production callers
	// leave them nil and Start() fills in the bound contracts.
	manager      keyperSetCounter
	keyBroadcast eonKeyBroadcast

	keyBroadcastCh chan *bindings.KeyBroadcastContractEonKeyBroadcast
}

func (s *EonPubKeySyncer) Start(ctx context.Context, runner service.Runner) error {
	if s.Handler == nil {
		return errors.New("no handler registered")
	}

	if s.manager == nil {
		s.manager = s.KeyperSetManager
	}
	if s.keyBroadcast == nil {
		s.keyBroadcast = s.KeyBroadcast
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
	subs, err := s.keyBroadcast.WatchEonKeyBroadcast(watchOpts, s.keyBroadcastCh)
	if err != nil {
		return err
	}
	runner.Go(func() error {
		err := s.watchNewEonPubkey(ctx, subs.Err())
		if err != nil {
			s.Log.Error("error watching new eon pubkey", err.Error())
		}
		subs.Unsubscribe()
		return err
	})
	return nil
}

// getInitialPubKeys performs the cold-start poll. It enumerates every keyper-set
// index (0..GetNumKeyperSets()-1) and delivers each *published* eon key to the
// handler, skipping indices whose key is not yet published rather than asking
// which set is active at the start block. This backfills the full history that
// the EonKeyBroadcast watch — which only ever delivers events at or after the
// start block — can never replay, and starts cleanly when no key is published
// yet. Eon keys are addressed by keyper-set index; activation is a downstream
// scheduling concern. See ADR 0010.
func (s *EonPubKeySyncer) getInitialPubKeys(ctx context.Context) ([]*event.EonPublicKey, error) {
	// This blocknumber specifies AT what state
	// the contract is called
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: s.StartBlock.Int,
	}
	if err := guardCallOpts(opts, false); err != nil {
		return nil, err
	}
	numKS, err := s.manager.GetNumKeyperSets(opts)
	if err != nil {
		return nil, makeCallError("GetNumKeyperSets", err)
	}

	initialPubKeys := make([]*event.EonPublicKey, 0, numKS)
	for i := uint64(0); i < numKS; i++ {
		e, err := s.GetEonPubKeyForEon(ctx, opts, i)
		if err != nil {
			return nil, err
		}
		// A nil event means the key at this index is not published yet; skip
		// it. Published keys arrive later via the watch. See ADR 0010.
		if e == nil {
			continue
		}
		initialPubKeys = append(initialPubKeys, e)
	}
	return initialPubKeys, nil
}

// GetEonPubKeyForEon reads the eon key at the given keyper-set index. It returns
// a nil event (and nil error) when the key is not yet published, detected by
// GetEonKey returning empty bytes — an unambiguous sentinel, since
// broadcastEonKey rejects zero-length keys. Callers treat nil as "no key yet"
// and skip the index. See ADR 0010.
func (s *EonPubKeySyncer) GetEonPubKeyForEon(ctx context.Context, opts *bind.CallOpts, eon uint64) (*event.EonPublicKey, error) {
	var err error
	opts, err = fixCallOpts(ctx, s.Client, opts)
	if err != nil {
		return nil, err
	}
	key, err := s.keyBroadcast.GetEonKey(opts, eon)
	if err != nil {
		return nil, makeCallError("GetEonKey", err)
	}
	if len(key) == 0 {
		return nil, nil
	}
	return &event.EonPublicKey{
		Eon:           eon,
		Key:           key,
		AtBlockNumber: number.BigToBlockNumber(opts.BlockNumber),
	}, nil
}

func (s *EonPubKeySyncer) watchNewEonPubkey(ctx context.Context, subsErr <-chan error) error {
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
		case err := <-subsErr:
			if err != nil {
				s.Log.Error("subscription error for watchNewEonPubkey", err.Error())
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
