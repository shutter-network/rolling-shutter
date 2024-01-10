package sync

import (
	"context"
	"crypto/ecdsa"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
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
	chainID *big.Int
	privKey *ecdsa.PrivateKey

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

func (s *ShutterL2Client) GetShutterState(ctx context.Context) (*event.ShutterState, error) {
	if s.sssync == nil {
		return nil, errors.Wrap(ErrServiceNotInstantiated, "ShutterStateSyncer service not instantiated")
	}
	opts := &bind.CallOpts{
		Context: ctx,
	}
	return s.sssync.GetShutterState(ctx, opts)
}

func (s *ShutterL2Client) GetKeyperSetByIndex(ctx context.Context, index uint64) (*event.KeyperSet, error) {
	if s.kssync == nil {
		return nil, errors.Wrap(ErrServiceNotInstantiated, "KeyperSetSyncer service not instantiated")
	}
	opts := &bind.CallOpts{
		Context: ctx,
	}
	return s.kssync.GetKeyperSetByIndex(ctx, opts, index)
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
		return nil, errors.Wrap(ErrServiceNotInstantiated, "ShutterStateSyncer service not instantiated")
	}
	opts := &bind.CallOpts{
		Context: ctx,
	}
	return s.epksync.GetEonPubKeyForEon(ctx, opts, eon)
}

func (s *ShutterL2Client) BroadcastEonKey(ctx context.Context, eon uint64, eonPubKey []byte) (*types.Transaction, error) {
	// TODO: first do a getEonKey. If we already have something (ideally the same)
	// don't do a transaction
	// s.KeyBroadcast.GetEonKey(eon)
	if s.privKey == nil {
		return nil, errors.New("can't broadcast eon public-key, client does not have a signer set")
	}
	chainID, err := s.ChainID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "retrieve chain id")
	}
	opts, err := bind.NewKeyedTransactorWithChainID(s.privKey, chainID)
	if err != nil {
		return nil, errors.Wrap(err, "construct signer transaction opts")
	}
	opts.Context = ctx
	return s.KeyBroadcast.BroadcastEonKey(opts, eon, eonPubKey)
}

// ChainID returns the chainid of the underlying L2 chain.
// This value is cached, since it is not expected to change.
func (s *ShutterL2Client) ChainID(ctx context.Context) (*big.Int, error) {
	if s.chainID == nil {
		cid, err := s.Client.ChainID(ctx)
		if err != nil {
			return nil, err
		}
		s.chainID = cid
	}
	return s.chainID, nil
}

func (s *ShutterL2Client) Start(_ context.Context, runner service.Runner) error {
	return runner.StartService(s.getServices()...)
}
