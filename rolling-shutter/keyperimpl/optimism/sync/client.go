package sync

import (
	"context"
	"io"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/shutter-network/shop-contracts/bindings"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/client"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/event"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/optimism/sync/syncer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/number"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/logger"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/service"
)

var noopLogger = &logger.NoopLogger{}

var ErrServiceNotInstantiated = errors.New("service is not instantiated, pass a handler function option.")

type ShutterSync interface {
	io.Closer
	// Start starts an additional worker syncing job
	Start() error
}

type ShutterL2Client struct {
	client.Client
	log log.Logger

	options *options

	KeyperSetManager *bindings.KeyperSetManager
	KeyBroadcast     *bindings.KeyBroadcastContract

	sssync  *syncer.ShutterStateSyncer
	kssync  *syncer.KeyperSetSyncer
	uhsync  *syncer.UnsafeHeadSyncer
	epksync *syncer.EonPubKeySyncer

	services []service.Service
}

func NewShutterL2Client(ctx context.Context, options ...Option) (*ShutterL2Client, error) {
	opts := defaultOptions()
	for _, option := range options {
		err := option(opts)
		if err != nil {
			return nil, err
		}
	}

	err := opts.verify()
	if err != nil {
		return nil, err
	}

	c := &ShutterL2Client{
		log:      noopLogger,
		services: []service.Service{},
	}
	err = opts.apply(ctx, c)
	if err != nil {
		return nil, err
	}
	c.options = opts
	return c, nil
}

func (s *ShutterL2Client) getServices() []service.Service {
	return s.services
}

//	func syncInitial() {
//		if s.ForceEmitActiveKeyperSet {
//			var b *number.BlockNumber
//			if s.StartBlock == nil {
//				b = number.LatestBlock
//			} else {
//				b = number.NewBlockNumber()
//				b.SetInt64(int64(*s.StartBlock))
//			}
//			activeKSAddred, err := s.getKeyperSetForBlock(ctx, b)
//			if err != nil {
//				return err
//			}
//			err = s.handler(activeKSAddred)
//			if err != nil {
//				return errors.Wrap(err, "handling of forced emission of active keyper set failed")
//			}
//		}
//	}

func (s *ShutterL2Client) GetShutterState(ctx context.Context) (*event.ShutterState, error) {
	if s.sssync == nil {
		return nil, errors.Wrap(ErrServiceNotInstantiated, "ShutterStateSyncer service not instantiated")
	}
	opts := &bind.CallOpts{
		Context: ctx,
	}
	return s.sssync.GetShutterState(ctx, opts)
}

func (s *ShutterL2Client) GetKeyperSetForBlock(ctx context.Context, b *number.BlockNumber) (*event.KeyperSet, error) {
	if s.kssync == nil {
		return nil, errors.Wrap(ErrServiceNotInstantiated, "KeyperSetSyncer service not instantiated")
	}
	opts := &bind.CallOpts{
		Context: ctx,
	}
	return s.kssync.GetKeyperSetForBlock(ctx, opts, b)
}

func (s *ShutterL2Client) GetEonPubKeyForEon(ctx context.Context, eon uint64) (*event.EonPublicKey, error) {
	if s.sssync == nil {
		return nil, errors.Wrap(ErrServiceNotInstantiated, "EonPubKeySyncer service not instantiated")
	}
	opts := &bind.CallOpts{
		Context: ctx,
	}
	return s.epksync.GetEonPubKeyForEon(ctx, opts, eon)
}

func (s *ShutterL2Client) Start(ctx context.Context, runner service.Runner) error {
	return runner.StartService(s.getServices()...)
}
