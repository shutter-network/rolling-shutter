package syncer

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/chainsync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
)

var _ ManualFilterHandler = &EonPubKeySyncer{}

type EonPubKeySyncer struct {
	Client           client.EthereumClient
	Log              log.Logger
	KeyBroadcast     *bindings.KeyBroadcastContract
	KeyperSetManager *bindings.KeyperSetManager
	Handler          event.EonPublicKeyHandler
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
		err := s.handle(ctx, iter.Event)
		if err != nil {
			s.Log.Error(
				"handler for `NewKeyperSet` errored",
				"error",
				err.Error(),
			)
		}
	}
	if err := iter.Error(); err != nil {
		return errors.Wrap(err, "filter iterator error")
	}
	return nil
}

func (s *EonPubKeySyncer) handle(ctx context.Context, newEonKey *bindings.KeyBroadcastContractEonKeyBroadcast) error {
	pubk := newEonKey.Key
	bn := newEonKey.Raw.BlockNumber
	ev := &event.EonPublicKey{
		Eon:           newEonKey.Eon,
		Key:           pubk,
		AtBlockNumber: number.NewBlockNumber(&bn),
	}
	err := s.Handler(ctx, ev)
	if err != nil {
		return err
	}
	return nil
}

func (s *EonPubKeySyncer) HandleVirtualEvent(ctx context.Context, block *number.BlockNumber) error {
	pubKs, err := s.getInitialPubKeys(ctx, block)
	if err != nil {
		return err
	}
	for _, k := range pubKs {
		err := s.Handler(ctx, k)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *EonPubKeySyncer) getInitialPubKeys(ctx context.Context, block *number.BlockNumber) ([]*event.EonPublicKey, error) {
	// This blocknumber specifies AT what state
	// the contract is called
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: block.Int,
	}
	numKS, err := s.KeyperSetManager.GetNumKeyperSets(opts)
	if err != nil {
		return nil, err
	}
	// this blocknumber specifies the argument to the contract
	// getter
	activeEon, err := s.KeyperSetManager.GetKeyperSetIndexByBlock(opts, block.Uint64())
	if err != nil {
		return nil, err
	}
	initialPubKeys := []*event.EonPublicKey{}
	// NOTE: These are pubkeys that at the state of s.StartBlock
	// are known to the contracts.
	// That way we recreate older broadcast publickey events.
	// We are only interested for keys that belong to keyper-set
	// that are currently active or will become active in
	// the future:
	for i := activeEon; i < numKS; i++ {
		e, err := s.GetEonPubKeyForEon(ctx, opts, i)
		if err != nil {
			return nil, err
		}
		// if e == nil, this means the keyperset did not broadcast a
		// key (yet)
		if e != nil {
			initialPubKeys = append(initialPubKeys, e)
		}
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
	if err != nil {
		return nil, err
	}
	// NOTE: Solidity returns the null value whenever
	// one tries to access a key in mapping that doesn't exist
	if len(key) == 0 {
		return nil, nil
	}
	return &event.EonPublicKey{
		Eon:           eon,
		Key:           key,
		AtBlockNumber: number.BigToBlockNumber(opts.BlockNumber),
	}, nil
}
